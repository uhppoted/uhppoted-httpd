package catalog

import (
	"fmt"
	"sync"
)

var catalog = map[string]record{}
var guard sync.Mutex

type record struct {
	ID      uint32
	deleted bool
}

func Put(deviceID uint32, oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog[oid] = record{
		ID:      deviceID,
		deleted: false,
	}
}

func Get(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range catalog {
			if !v.deleted && v.ID == deviceID {
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

		catalog[oid] = record{
			ID:      deviceID,
			deleted: false,
		}

		return oid
	}
}

func Delete(oid string) {
	guard.Lock()
	defer guard.Unlock()

	if v, ok := catalog[oid]; ok {
		catalog[oid] = record{
			ID:      v.ID,
			deleted: true,
		}
	}
}

func Find(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range catalog {
			if !v.deleted && v.ID == deviceID {
				return oid
			}
		}
	}

	return ""
}
