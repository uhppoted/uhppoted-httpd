package memdb

import (
	"fmt"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type catalog struct {
	interfaces  table[*entry]
	controllers controllers
	doors       table[*entry]
	cards       table[*entry]
	groups      table[*entry]
	events      table[*entry]
	logs        table[*entry]
	users       table[*entry]
	guard       sync.RWMutex
}

var db = catalog{
	controllers: controllers{
		base:  schema.ControllersOID,
		m:     map[schema.OID]controller{},
		limit: 32,
	},

	interfaces: table[*entry]{
		base:  schema.InterfacesOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	doors: table[*entry]{
		base:  schema.DoorsOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	cards: table[*entry]{
		base:  schema.CardsOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	groups: table[*entry]{
		base:  schema.GroupsOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	events: table[*entry]{
		base:  schema.EventsOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	logs: table[*entry]{
		base:  schema.LogsOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
	users: table[*entry]{
		base:  schema.UsersOID,
		m:     map[schema.OID]*entry{},
		limit: 32,
	},
}

func Catalog() *catalog {
	return &db
}

func (cc *catalog) Clear() {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	cc.interfaces.Clear()
	cc.controllers.Clear()
	cc.doors.Clear()
	cc.cards.Clear()
	cc.groups.Clear()
	cc.events.Clear()
	cc.logs.Clear()
	cc.users.Clear()

	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache = map[schema.OID]value{}
}

func (cc *catalog) NewT(t ctypes.Type, v interface{}) schema.OID {
	if t == ctypes.TController {
		cc.guard.Lock()
		defer cc.guard.Unlock()

		if deviceID := v.(uint32); deviceID != 0 {
			for oid, v := range cc.controllers.m {
				if !v.deleted && v.ID == deviceID {
					return oid
				}
			}
		}

		return cc.controllers.New(v.(uint32))
	}

	// NTS: only support a single interface at this point in time
	if t == ctypes.TInterface {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	m, ok := cc.tableFor(t)
	if !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	return m.New(v)
}

func (cc *catalog) PutT(t ctypes.Type, v interface{}, oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	if t == ctypes.TController {
		cc.controllers.Put(oid, v.(uint32))
		return
	}

	if m, ok := cc.tableFor(t); !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	} else {
		m.Put(oid, v)
	}
}

func (cc *catalog) DeleteT(t ctypes.Type, oid schema.OID) {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	switch t {
	case ctypes.TController:
		cc.controllers.Delete(oid)

	default:
		if tt, ok := cc.tableFor(t); ok {
			tt.Delete(oid)
		}
	}
}

func (cc *catalog) ListT(t ctypes.Type) []schema.OID {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	list := []schema.OID{}

	switch t {
	case ctypes.TController:
		for d, v := range cc.controllers.m {
			if !v.deleted {
				list = append(list, d)
			}
		}

	default:
		if tt, ok := cc.tableFor(t); ok {
			for d, v := range tt.m {
				if !v.deleted {
					list = append(list, d)
				}
			}
		}
	}

	return list
}

func (cc *catalog) HasT(t ctypes.Type, oid schema.OID) bool {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	switch t {
	case ctypes.TController:
		if v, ok := cc.controllers.m[oid]; ok && !v.deleted {
			return true
		}

	default:
		if tt, ok := cc.tableFor(t); ok {
			if v, ok := tt.m[oid]; ok && !v.deleted {
				return true
			}
		}
	}

	return false
}

func (cc *catalog) FindController(deviceID uint32) schema.OID {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	if deviceID != 0 {
		for oid, v := range cc.controllers.m {
			if v.ID == deviceID && !v.deleted {
				return oid
			}
		}
	}

	return ""
}

func (cc *catalog) tableFor(t ctypes.Type) (table[*entry], bool) {
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
		return table[*entry]{}, false
	}
}
