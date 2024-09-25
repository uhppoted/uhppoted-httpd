package auth

import (
	"strings"
	"sync"
	"time"
)

type cache struct {
	canView map[string]result
	guard   sync.RWMutex
	touched time.Time
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

func cacheCanView(uid string, operant Operant, field string, f func() (result, error)) (result, error) {
	key := operant.CacheKey()

	if uid != "" && key != "" {
		v, _ := _cache.LoadOrStore(uid, &cache{
			canView: map[string]result{},
			guard:   sync.RWMutex{},
			touched: time.Now(),
		})

		if c, ok := v.(*cache); ok {
			c.touched = time.Now()

			// ... make cache key
			var b strings.Builder

			b.WriteString(key)

			if field != "" {
				b.WriteString("::")
				b.WriteString(field)
			}

			k := b.String()

			// ... get cache entry for operant+field
			if rs, ok := cacheGet(c, k); ok {
				return rs, nil
			} else if rs, err := f(); err != nil {
				return rs, err
			} else {
				return cacheTryPut(c, k, rs), nil
			}
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
