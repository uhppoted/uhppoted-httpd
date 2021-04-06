package catalog

import (
	"fmt"
	"sync"
)

var catalog = struct {
	interfaces  map[string]struct{}
	controllers map[string]record
}{
	interfaces:  map[string]struct{}{},
	controllers: map[string]record{},
}

var guard sync.Mutex

type record struct {
	ID      uint32
	deleted bool
}

func PutInterface(oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog.interfaces[oid] = struct{}{}
}

func PutController(deviceID uint32, oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog.controllers[oid] = record{
		ID:      deviceID,
		deleted: false,
	}
}

func Get(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range catalog.controllers {
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
		for v, _ := range catalog.controllers {
			if v == oid {
				continue loop
			}
		}

		catalog.controllers[oid] = record{
			ID:      deviceID,
			deleted: false,
		}

		return oid
	}
}

func Delete(oid string) {
	guard.Lock()
	defer guard.Unlock()

	if v, ok := catalog.controllers[oid]; ok {
		catalog.controllers[oid] = record{
			ID:      v.ID,
			deleted: true,
		}
	}
}

func Find(deviceID uint32) string {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range catalog.controllers {
			if !v.deleted && v.ID == deviceID {
				return oid
			}
		}
	}

	return ""
}
