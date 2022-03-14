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

type controller struct {
	ID      uint32
	deleted bool
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

var baseOIDs = map[ctypes.Type]schema.OID{
	ctypes.TInterface:  schema.InterfacesOID,
	ctypes.TController: schema.ControllersOID,
	ctypes.TDoor:       schema.DoorsOID,
	ctypes.TCard:       schema.CardsOID,
	ctypes.TGroup:      schema.GroupsOID,
	ctypes.TEvent:      schema.EventsOID,
	ctypes.TLog:        schema.LogsOID,
	ctypes.TUser:       schema.UsersOID,
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
	if t == ctypes.TInterface {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	if t == ctypes.TController {
		return cc.newController(v.(uint32))
	}

	base, ok := baseOIDs[t]
	if !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	m, ok := cc.mapFor(t)
	if !ok {
		panic(fmt.Sprintf("Unsupported base OID (%v)", base))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

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

	if t == ctypes.TController {
		cc.controllers[oid] = controller{
			ID:      v.(uint32),
			deleted: false,
		}

		return
	}

	if m, ok := cc.mapFor(t); !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	} else {
		m[oid] = struct{}{}
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

func (cc *catalog) mapFor(t ctypes.Type) (map[schema.OID]struct{}, bool) {
	switch t {
	case ctypes.TInterface:
		return cc.interfaces, true

	case ctypes.TDoor:
		return cc.doors, true

	case ctypes.TCard:
		return cc.cards, true

	case ctypes.TGroup:
		return cc.groups, true

	case ctypes.TEvent:
		return cc.events, true

	case ctypes.TLog:
		return cc.logs, true

	case ctypes.TUser:
		return cc.users, true

	default:
		return nil, false
	}
}
