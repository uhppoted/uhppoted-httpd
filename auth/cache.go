package auth

import (
	"strings"
	"sync"
	"time"
)

type cache struct {
	touched time.Time
	canView map[string]result
}

var _cache = sync.Map{}
var idleTime = 300 * time.Second

func init() {
	tick := time.Tick(30 * time.Second)

	go func() {
		for range tick {
			_cache.Range(func(k, v any) bool {
				if dt := time.Since(v.(cache).touched); dt > idleTime {
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
		v, _ := _cache.LoadOrStore(uid, cache{
			touched: time.Now(),
			canView: map[string]result{},
		})

		if c, ok := v.(cache); ok {
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
			if rs, ok := c.canView[k]; ok {
				return rs, nil
			} else if rs, err := f(); err != nil {
				return rs, err
			} else {
				c.canView[k] = rs

				return rs, nil
			}
		}
	}

	return f()
}
