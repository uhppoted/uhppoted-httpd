package memdb

import (
	"fmt"
	"sync"

	cat "github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type catalog struct {
	interfaces  Table
	controllers Table
	doors       Table
	cards       Table
	groups      Table
	events      Table
	logs        Table
	users       Table
	guard       sync.RWMutex
}

var db = catalog{
	controllers: &controllers{
		base: schema.ControllersOID,
		m:    map[schema.OID]*controller{},
	},

	interfaces: &table{
		base: schema.InterfacesOID,
		m:    map[schema.OID]*record{},
	},
	doors: &table{
		base: schema.DoorsOID,
		m:    map[schema.OID]*record{},
	},
	cards: &table{
		base: schema.CardsOID,
		m:    map[schema.OID]*record{},
	},
	groups: &table{
		base: schema.GroupsOID,
		m:    map[schema.OID]*record{},
	},
	events: &table{
		base: schema.EventsOID,
		m:    map[schema.OID]*record{},
	},
	logs: &table{
		base: schema.LogsOID,
		m:    map[schema.OID]*record{},
	},
	users: &table{
		base: schema.UsersOID,
		m:    map[schema.OID]*record{},
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

		u := v.(cat.CatalogController)

		if deviceID := u.DeviceID; deviceID != 0 {
			for oid, c := range cc.controllers.(*controllers).m {
				if !c.deleted && c.ID == deviceID {
					return oid
				}
			}
		}

		return cc.controllers.New(u)
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
		cc.controllers.Put(oid, v.(cat.CatalogController).DeviceID)
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
		for d, v := range cc.controllers.(*controllers).m {
			if !v.deleted {
				list = append(list, d)
			}
		}

	default:
		if tt, ok := cc.tableFor(t); ok {
			for d, v := range tt.(*table).m {
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
		if v, ok := cc.controllers.(*controllers).m[oid]; ok && !v.deleted {
			return true
		}

	default:
		if tt, ok := cc.tableFor(t); ok {
			if v, ok := tt.(*table).m[oid]; ok && !v.deleted {
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
		for oid, v := range cc.controllers.(*controllers).m {
			if v.ID == deviceID && !v.deleted {
				return oid
			}
		}
	}

	return ""
}

func (cc *catalog) tableFor(t ctypes.Type) (Table, bool) {
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

// func tableFor[T ctypes.CatalogType](v T) Table {
// 	t := fmt.Sprintf("%T", v)
// 	switch t {
// 	case "ctypes.CatalogInterface":
// 		return cc.interfaces
//
// 	case "ctypes.CatalogController":
// 		return cc.controllers
//
// 	case "ctypes.CatalogCard":
// 		return cc.cards
//
// 	case "ctypes.CatalogDoor":
// 		return cc.doors
//
// 	case "ctypes.CatalogGroup":
// 		return cc.groups
//
// 	case "ctypes.CatalogEvent":
// 		return cc.events
//
// 	case "ctypes.CatalogLogEntry":
// 		return cc.logs
//
// 	case "ctypes.CatalogUser":
// 		return cc.users
//
// 	default:
// 		return nil
// 	}
// }
