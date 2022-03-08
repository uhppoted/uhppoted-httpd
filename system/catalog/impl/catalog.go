package memdb

import (
	"fmt"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type catalog struct {
	interfaces  map[schema.OID]struct{}
	controllers map[schema.OID]controller
	doors       map[schema.OID]struct{}
	cards       map[schema.OID]struct{}
	groups      map[schema.OID]struct{}
	events      map[schema.OID]struct{}
	logs        map[schema.OID]struct{}
	users       map[schema.OID]struct{}
	guard       sync.Mutex
}

var db = catalog{
	interfaces:  map[schema.OID]struct{}{},
	controllers: map[schema.OID]controller{},
	doors:       map[schema.OID]struct{}{},
	cards:       map[schema.OID]struct{}{},
	groups:      map[schema.OID]struct{}{},
	events:      map[schema.OID]struct{}{},
	logs:        map[schema.OID]struct{}{},
	users:       map[schema.OID]struct{}{},
}

type controller struct {
	ID      uint32
	deleted bool
}

func Catalog() *catalog {
	return &db
}

func (cc *catalog) Doors() map[schema.OID]struct{} {
	return cc.doors
}

func (cc *catalog) Groups() map[schema.OID]struct{} {
	return cc.groups
}

func (cc *catalog) Clear() {
	cc.interfaces = map[schema.OID]struct{}{}
	cc.controllers = map[schema.OID]controller{}
	cc.doors = map[schema.OID]struct{}{}
	cc.cards = map[schema.OID]struct{}{}
	cc.groups = map[schema.OID]struct{}{}
	cc.events = map[schema.OID]struct{}{}
	cc.logs = map[schema.OID]struct{}{}

	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache = map[schema.OID]value{}
}

func (cc *catalog) PutInterface(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.interfaces[oid] = struct{}{}
}

func (cc *catalog) PutController(deviceID uint32, oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.controllers[oid] = controller{
		ID:      deviceID,
		deleted: false,
	}
}

func (cc *catalog) PutDoor(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.doors[oid] = struct{}{}
}

func (cc *catalog) PutCard(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.cards[oid] = struct{}{}
}

func (cc *catalog) PutGroup(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.groups[oid] = struct{}{}
}

func (cc *catalog) HasGroup(oid schema.OID) bool {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	_, ok := cc.groups[oid]

	return ok
}

func (cc *catalog) PutEvent(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.events[oid] = struct{}{}
}

func (cc *catalog) PutLogEntry(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.logs[oid] = struct{}{}
}

func (cc *catalog) PutUser(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.users[oid] = struct{}{}
}

func (cc *catalog) NewController(deviceID uint32) schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	if deviceID != 0 {
		for oid, v := range cc.controllers {
			if !v.deleted && v.ID == deviceID {
				return oid
			}
		}
	}

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.ControllersOID, item))
		for v, _ := range cc.controllers {
			if v == oid {
				continue loop
			}
		}

		cc.controllers[oid] = controller{
			ID:      deviceID,
			deleted: false,
		}

		return oid
	}
}

func (cc *catalog) NewDoor() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.DoorsOID, item))
		for v, _ := range cc.doors {
			if v == oid {
				continue loop
			}
		}

		cc.doors[oid] = struct{}{}
		return oid
	}
}

func (cc *catalog) NewCard() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.CardsOID, item))
		for v, _ := range cc.cards {
			if v == oid {
				continue loop
			}
		}

		cc.cards[oid] = struct{}{}

		return oid
	}
}

func (cc *catalog) NewGroup() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.GroupsOID, item))
		for v, _ := range cc.groups {
			if v == oid {
				continue loop
			}
		}

		cc.groups[oid] = struct{}{}

		return oid
	}
}

func (cc *catalog) NewEvent() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.EventsOID, item))
		for v, _ := range cc.events {
			if v == oid {
				continue loop
			}
		}

		cc.events[oid] = struct{}{}

		return oid
	}
}

func (cc *catalog) NewLogEntry() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.LogsOID, item))
		for v, _ := range cc.logs {
			if v == oid {
				continue loop
			}
		}

		cc.logs[oid] = struct{}{}

		return oid
	}
}

func (cc *catalog) NewUser() schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.UsersOID, item))
		for v, _ := range cc.users {
			if v == oid {
				continue loop
			}
		}

		cc.users[oid] = struct{}{}

		return oid
	}
}

func (cc *catalog) Delete(oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	if v, ok := cc.controllers[oid]; ok {
		cc.controllers[oid] = controller{
			ID:      v.ID,
			deleted: true,
		}
	}
}

func (cc *catalog) FindController(deviceID uint32) schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	if deviceID != 0 {
		for oid, v := range cc.controllers {
			if v.ID == deviceID && !v.deleted {
				return oid
			}
		}
	}

	return ""
}
