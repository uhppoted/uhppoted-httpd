package memdb

import (
	"fmt"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
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

func (cc *catalog) NewT(t ctypes.Type, v interface{}) schema.OID {
	switch t {
	//	case ctypes.TInterface:
	//		cc.interfaces[oid] = struct{}{}

	case ctypes.TController:
		return cc.newController(v.(uint32))

	case ctypes.TDoor:
		return cc.newOID(schema.DoorsOID)

	case ctypes.TCard:
		return cc.newOID(schema.CardsOID)

	case ctypes.TGroup:
		return cc.newOID(schema.GroupsOID)

	case ctypes.TEvent:
		return cc.newOID(schema.EventsOID)

	case ctypes.TLog:
		return cc.newOID(schema.LogsOID)

	case ctypes.TUser:
		return cc.newOID(schema.UsersOID)

	default:
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}
}

func (cc *catalog) newOID(base schema.OID) schema.OID {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	var m map[schema.OID]struct{}

	switch base {
	case schema.DoorsOID:
		m = cc.doors

	case schema.CardsOID:
		m = cc.cards

	case schema.GroupsOID:
		m = cc.groups

	case schema.EventsOID:
		m = cc.events

	case schema.LogsOID:
		m = cc.logs

	case schema.UsersOID:
		m = cc.users

	default:
		panic(fmt.Sprintf("Unsupported base OID (%v)", base))
	}

	item := 0
loop:
	for {
		item += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", base, item))
		for v, _ := range m {
			if v == oid {
				continue loop
			}
		}

		m[oid] = struct{}{}
		return oid
	}
}

func (cc *catalog) newController(deviceID uint32) schema.OID {
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

func (cc *catalog) PutT(t ctypes.Type, v interface{}, oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	switch t {
	case ctypes.TInterface:
		cc.interfaces[oid] = struct{}{}

	case ctypes.TController:
		cc.controllers[oid] = controller{
			ID:      v.(uint32),
			deleted: false,
		}

	case ctypes.TDoor:
		cc.doors[oid] = struct{}{}

	case ctypes.TCard:
		cc.cards[oid] = struct{}{}

	case ctypes.TGroup:
		cc.groups[oid] = struct{}{}

	case ctypes.TEvent:
		cc.events[oid] = struct{}{}

	case ctypes.TLog:
		cc.logs[oid] = struct{}{}

	case ctypes.TUser:
		cc.users[oid] = struct{}{}

	default:
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
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

func (cc *catalog) Doors() map[schema.OID]struct{} {
	return cc.doors
}

func (cc *catalog) Groups() map[schema.OID]struct{} {
	return cc.groups
}

func (cc *catalog) HasGroup(oid schema.OID) bool {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	_, ok := cc.groups[oid]

	return ok
}
