package auth

import (
	"time"
)

type cache struct {
	// uid     string
	touched time.Time
}

var _cache = map[string]cache{}
var idleTime = 30 * time.Second

func init() {
	c := time.Tick(15 * time.Second)

	go func() {
		for range c {
			idle := []string{}
			for k, v := range _cache {
				if dt := time.Since(v.touched); dt > idleTime {
					idle = append(idle, k)
				}
			}

			for _, uid := range idle {
				infof("AUTH", "clearing grules cached for idle uid %v", uid)
			}
		}
	}()
}

func cacheClear() {
	infof("AUTH", "cleared grules cache")
}

func cacheCanView(uid string, operant Operant, field string) (result, bool) {
	return result{}, false
}
