package auth

import (
	"strings"
	"sync"
	"time"
)

type cache struct {
	canView  map[string]result
	canCache map[string]bool
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

	// ... get cache for uid:role
	cacheId := strings.Join([]string{a.uid, a.role}, ":")
	v, _ := _cache.LoadOrStore(cacheId, &cache{
		canView:  map[string]result{},
		canCache: map[string]bool{},
		guard:    sync.RWMutex{},
		touched:  time.Now(),
	})

	if c, ok := v.(*cache); ok {
		c.touched = time.Now()

		// ... cacheable?
		if _, ok := c.canCache[key]; !ok {
			if err := CanCache(a, operant, field, "CanView", rulesets...); err != nil {
				c.canCache[key] = false
			} else {
				c.canCache[key] = true
			}
		}

		cacheable := c.canCache[key]

		// ... get cache entry for operant+field
		if rs, ok := cacheGet(c, key); ok {
			return rs, nil
		} else if rs, err := f(); err != nil {
			return rs, err
		} else if cacheable {
			return cacheTryPut(c, key, rs), nil
		} else {
			return rs, nil
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
