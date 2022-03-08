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

var guard sync.Mutex

type controller struct {
	ID      uint32
	deleted bool
}

func Catalog() *catalog {
	return &db
}

func (c *catalog) Doors() map[schema.OID]struct{} {
	return c.doors
}

func (c *catalog) Groups() map[schema.OID]struct{} {
	return c.groups
}

func (c *catalog) Clear() {
	c.interfaces = map[schema.OID]struct{}{}
	c.controllers = map[schema.OID]controller{}
	c.doors = map[schema.OID]struct{}{}
	c.cards = map[schema.OID]struct{}{}
	c.groups = map[schema.OID]struct{}{}
	c.events = map[schema.OID]struct{}{}
	c.logs = map[schema.OID]struct{}{}

	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache = map[schema.OID]value{}
}

func (c *catalog) PutInterface(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.interfaces[oid] = struct{}{}
}

func (c *catalog) PutController(deviceID uint32, oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.controllers[oid] = controller{
		ID:      deviceID,
		deleted: false,
	}
}

func (c *catalog) PutDoor(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.doors[oid] = struct{}{}
}

func (c *catalog) PutCard(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.cards[oid] = struct{}{}
}

func (c *catalog) PutGroup(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.groups[oid] = struct{}{}
}

func (c *catalog) HasGroup(oid schema.OID) bool {
	guard.Lock()
	defer guard.Unlock()

	_, ok := c.groups[oid]

	return ok
}

func (c *catalog) PutEvent(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.events[oid] = struct{}{}
}

func (c *catalog) PutLogEntry(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.logs[oid] = struct{}{}
}

func (c *catalog) PutUser(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	c.users[oid] = struct{}{}
}

func (c *catalog) NewController(deviceID uint32) schema.OID {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range c.controllers {
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
		for v, _ := range c.controllers {
			if v == oid {
				continue loop
			}
		}

		c.controllers[oid] = controller{
			ID:      deviceID,
			deleted: false,
		}

		return oid
	}
}

func (c *catalog) NewDoor() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.DoorsOID, item))
		for v, _ := range c.doors {
			if v == oid {
				continue loop
			}
		}

		c.doors[oid] = struct{}{}
		return oid
	}
}

func (c *catalog) NewCard() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.CardsOID, item))
		for v, _ := range c.cards {
			if v == oid {
				continue loop
			}
		}

		c.cards[oid] = struct{}{}

		return oid
	}
}

func (c *catalog) NewGroup() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.GroupsOID, item))
		for v, _ := range c.groups {
			if v == oid {
				continue loop
			}
		}

		c.groups[oid] = struct{}{}

		return oid
	}
}

func (c *catalog) NewEvent() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.EventsOID, item))
		for v, _ := range c.events {
			if v == oid {
				continue loop
			}
		}

		c.events[oid] = struct{}{}

		return oid
	}
}

func (c *catalog) NewLogEntry() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.LogsOID, item))
		for v, _ := range c.logs {
			if v == oid {
				continue loop
			}
		}

		c.logs[oid] = struct{}{}

		return oid
	}
}

func (c *catalog) NewUser() schema.OID {
	guard.Lock()
	defer guard.Unlock()

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", schema.UsersOID, item))
		for v, _ := range c.users {
			if v == oid {
				continue loop
			}
		}

		c.users[oid] = struct{}{}

		return oid
	}
}

func (c *catalog) Delete(oid schema.OID) {
	guard.Lock()
	defer guard.Unlock()

	if v, ok := db.controllers[oid]; ok {
		c.controllers[oid] = controller{
			ID:      v.ID,
			deleted: true,
		}
	}
}

func (c *catalog) FindController(deviceID uint32) schema.OID {
	guard.Lock()
	defer guard.Unlock()

	if deviceID != 0 {
		for oid, v := range c.controllers {
			if v.ID == deviceID && !v.deleted {
				return oid
			}
		}
	}

	return ""
}
