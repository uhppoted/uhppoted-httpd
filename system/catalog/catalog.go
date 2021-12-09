package catalog

import (
	"fmt"
	"sync"
)

var catalog = struct {
	interfaces  map[OID]struct{}
	controllers map[OID]controller
	doors       map[OID]struct{}
	cards       map[OID]struct{}
	groups      map[OID]struct{}
	events      map[OID]struct{}
	logs        map[OID]struct{}
}{
	interfaces:  map[OID]struct{}{},
	controllers: map[OID]controller{},
	doors:       map[OID]struct{}{},
	cards:       map[OID]struct{}{},
	groups:      map[OID]struct{}{},
	events:      map[OID]struct{}{},
	logs:        map[OID]struct{}{},
}

var guard sync.Mutex

type controller struct {
	ID      uint32
	deleted bool
}

func Clear() {
	catalog.interfaces = map[OID]struct{}{}
	catalog.controllers = map[OID]controller{}
	catalog.doors = map[OID]struct{}{}
	catalog.cards = map[OID]struct{}{}
	catalog.groups = map[OID]struct{}{}
	catalog.events = map[OID]struct{}{}
	catalog.logs = map[OID]struct{}{}

	cache.guard.Lock()
	defer cache.guard.Unlock()
	cache.cache = map[OID]value{}
}

func PutInterface(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.interfaces[oid] = struct{}{}
}

func PutController(deviceID uint32, oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.controllers[oid] = controller{
		ID:      deviceID,
		deleted: false,
	}
}

func PutDoor(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.doors[oid] = struct{}{}
}

func PutCard(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.cards[oid] = struct{}{}
}

func PutGroup(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.groups[oid] = struct{}{}
}

func HasGroup(oid OID) bool {
	guard.Lock()
	defer guard.Unlock()

	_, ok := catalog.groups[oid]

	return ok
}

func PutEvent(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.events[oid] = struct{}{}
}

func PutLogEntry(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	catalog.logs[oid] = struct{}{}
}

func NewController(deviceID uint32) OID {
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
		oid := OID(fmt.Sprintf("%v.%d", ControllersOID, item))
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

func NewDoor() OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := OID(fmt.Sprintf("%v.%d", DoorsOID, item))
		for v, _ := range catalog.doors {
			if v == oid {
				continue loop
			}
		}

		catalog.doors[oid] = struct{}{}
		return oid
	}
}

func NewCard() OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := OID(fmt.Sprintf("%v.%d", CardsOID, item))
		for v, _ := range catalog.cards {
			if v == oid {
				continue loop
			}
		}

		catalog.cards[oid] = struct{}{}

		return oid
	}
}

func NewGroup() OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := OID(fmt.Sprintf("%v.%d", GroupsOID, item))
		for v, _ := range catalog.groups {
			if v == oid {
				continue loop
			}
		}

		catalog.groups[oid] = struct{}{}

		return oid
	}
}

func NewEvent() OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := OID(fmt.Sprintf("%v.%d", GroupsOID, item))
		for v, _ := range catalog.events {
			if v == oid {
				continue loop
			}
		}

		catalog.events[oid] = struct{}{}

		return oid
	}
}

func NewLogEntry() OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := OID(fmt.Sprintf("%v.%d", LogsOID, item))
		for v, _ := range catalog.logs {
			if v == oid {
				continue loop
			}
		}

		catalog.logs[oid] = struct{}{}

		return oid
	}
}

func Delete(oid OID) {
	guard.Lock()
	defer guard.Unlock()

	if v, ok := catalog.controllers[oid]; ok {
		catalog.controllers[oid] = controller{
			ID:      v.ID,
			deleted: true,
		}
	}
}

func FindController(deviceID uint32) OID {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range catalog.controllers {
			if v.ID == deviceID && !v.deleted {
				return oid
			}
		}
	}

	return ""
}
