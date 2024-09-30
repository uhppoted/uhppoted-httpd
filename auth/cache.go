package auth

import (
	"strings"
	"sync"
	"time"
)

type cache struct {
	canView  map[string]result
	canCache sync.Map
	guard    sync.RWMutex
	touched  time.Time
}

var _cache = sync.Map{}
var idleTime = 300 * time.Second

func init() {
	tick := time.Tick(30 * time.Second)

	go func() {
		for range tick {
			_cache.Range(func(k, v any) bool {
				if dt := time.Since(v.(*cache).touched); dt > idleTime {
					_cache.Delete(k)
					infof("AUTH", "cleared cached grules for '%v'", k)
				}

				return true
			})
		}
	}()
}

func cacheClear() {
	_cache.Clear()

	infof("AUTH", "cleared grules cache")
}

// NTS: slightly weird caching logic spreads expensive grules evaluations over two invocations
//
//	for the relatively minor cost of possibly unnecessarily caching a 'can view' result
func cacheCanView(a *authorizator, operant Operant, field string, f func() (result, error), rulesets ...RuleSet) (result, error) {
	if a == nil || a.uid == "" {
		return f()
	}

	// ... make cache key
	key := operant.CacheKey()
	if key == "" {
		return f()
	} else if field != "" {
		key = strings.Join([]string{key, field}, "::")
	}

	cacheable := func(c *cache) bool {
		if v, ok := c.canCache.Load(key); ok {
			return v.(bool)
		}

		if err := CanCache(a, operant, field, "CanView", rulesets...); err != nil {
			c.canCache.Store(key, false)
			return false
		}

		c.canCache.Store(key, true)
		return true
	}

	// ... get cache for uid:role
	cacheId := strings.Join([]string{a.uid, a.role}, ":")
	v, _ := _cache.LoadOrStore(cacheId, &cache{
		canView:  map[string]result{},
		canCache: sync.Map{},
		guard:    sync.RWMutex{},
		touched:  time.Now(),
	})

	if c, ok := v.(*cache); ok {
		c.touched = time.Now()

		if rs, ok := cacheGet(c, key); ok {
			if cacheable(c) {
				return rs, nil
			}
		} else if rs, err := f(); err != nil {
			return rs, err
		} else {
			return cacheTryPut(c, key, rs), nil
		}
	}

	return f()
}

func cacheGet(c *cache, key string) (result, bool) {
	c.guard.RLock()
	defer c.guard.RUnlock()

	rs, ok := c.canView[key]

	return rs, ok
}

// NTS: one of the rare occasions when TryLock is actually justified
func cacheTryPut(c *cache, key string, rs result) result {
	if c.guard.TryLock() {
		defer c.guard.Unlock()

		c.canView[key] = rs
	}

	return rs
}
