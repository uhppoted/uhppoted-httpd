package auth

import (
	"strings"
	"time"
)

type cache struct {
	touched time.Time
	canView map[string]result
}

var _cache = map[string]cache{}
var idleTime = 300 * time.Second

func init() {
	c := time.Tick(30 * time.Second)

	go func() {
		for range c {
			idle := []string{}
			for k, v := range _cache {
				if dt := time.Since(v.touched); dt > idleTime {
					idle = append(idle, k)
				}
			}

			for _, uid := range idle {
				delete(_cache, uid)
				infof("AUTH", "cleared cached grules for '%v'", uid)
			}
		}
	}()
}

func cacheClear() {
	_cache = map[string]cache{}

	infof("AUTH", "cleared grules cache")
}

func cacheCanView(uid string, operant Operant, field string, f func() (result, error)) (result, error) {
	key := operant.CacheKey()

	if uid != "" && key != "" {
		// ... get cache for UID
		if _, ok := _cache[uid]; !ok {
			_cache[uid] = cache{
				touched: time.Now(),
				canView: map[string]result{},
			}
		}

		c := _cache[uid]
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

	return f()
}
