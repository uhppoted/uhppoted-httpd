package catalog

import (
	"fmt"
	"sync"
)

var catalog = map[string]uint32{}
var guard sync.Mutex

func Put(deviceID uint32, oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog[oid] = deviceID
}

func Get(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, id := range catalog {
			if id == deviceID {
				return oid
			}
		}
	}

	item := 0
loop:
	for {
		item += 1
		oid := fmt.Sprintf("0.1.1.%d", item)
		for v, _ := range catalog {
			if v == oid {
				continue loop
			}
		}

		catalog[oid] = deviceID

		return oid
	}
}

func Find(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, id := range catalog {
			if id == deviceID {
				return oid
			}
		}
	}

	return ""
}
