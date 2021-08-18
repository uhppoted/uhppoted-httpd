package catalog

import (
	"fmt"
	"sync"
)

type Object struct {
	OID   string `json:"OID"`
	Value string `json:"value"`
}

var catalog = struct {
	interfaces  map[string]struct{}
	controllers map[string]controller
	doors       map[string]struct{}
}{
	interfaces:  map[string]struct{}{},
	controllers: map[string]controller{},
	doors:       map[string]struct{}{},
}

var guard sync.Mutex

type controller struct {
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

	catalog.controllers[oid] = controller{
		ID:      deviceID,
		deleted: false,
	}
}

func PutDoor(oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog.doors[oid] = struct{}{}
}

func GetController(deviceID uint32) string {
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
		oid := fmt.Sprintf("0.1.1.2.%d", item)
		for v, _ := range catalog.controllers {
			if v == oid {
				continue loop
			}
		}

		catalog.controllers[oid] = controller{
			ID:      deviceID,
			deleted: false,
		}

		return oid
	}
}

func NewDoor() string {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := fmt.Sprintf("0.3.%d", item)
		for v, _ := range catalog.doors {
			if v == oid {
				continue loop
			}
		}

		catalog.doors[oid] = struct{}{}
		return oid
	}
}

func Delete(oid string) {
	guard.Lock()
	defer guard.Unlock()

	if v, ok := catalog.controllers[oid]; ok {
		catalog.controllers[oid] = controller{
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
