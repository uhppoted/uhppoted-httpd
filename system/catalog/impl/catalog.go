package memdb

import (
	"fmt"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type db struct {
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

var dbx = db{
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

func Catalog() *db {
	return &dbx
}

func (cc *db) Clear() {
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

func (cc *db) NewT(v any) schema.OID {
	t := TypeOf(v)
	if t == TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	if t == TController {
		cc.guard.Lock()
		defer cc.guard.Unlock()

		u := v.(catalog.CatalogController)

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
	if t == TInterface {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	m, ok := cc.tableForT(t)
	if !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	return m.New(v)
}

func (cc *db) PutT(v any, oid schema.OID) {
	t := TypeOf(v)
	if t == TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	if t == TController {
		cc.controllers.Put(oid, v.(catalog.CatalogController).DeviceID)
		return
	}

	if m, ok := cc.tableForT(t); !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	} else {
		m.Put(oid, v)
	}
}

func (cc *db) DeleteT(v any, oid schema.OID) {
	t := TypeOf(v)
	if t == TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	switch t {
	case TController:
		cc.controllers.Delete(oid)

	default:
		if tt, ok := cc.tableForT(t); ok {
			tt.Delete(oid)
		}
	}
}

func (cc *db) ListT(oid schema.OID) []schema.OID {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	list := []schema.OID{}

	if tt, ok := cc.tableFor(oid); ok {
		for d, v := range tt.(*table).m {
			if !v.deleted {
				list = append(list, d)
			}
		}
	}

	return list
}

func (cc *db) HasT(v any, oid schema.OID) bool {
	t := TypeOf(v)
	if t == TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.RLock()
	defer cc.guard.RUnlock()

	switch t {
	case TController:
		if v, ok := cc.controllers.(*controllers).m[oid]; ok && !v.deleted {
			return true
		}

	default:
		if tt, ok := cc.tableForT(t); ok {
			if v, ok := tt.(*table).m[oid]; ok && !v.deleted {
				return true
			}
		}
	}

	return false
}

func (cc *db) FindController(deviceID uint32) schema.OID {
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

// TODO Remove, pending migration to real Go generics
func (cc *db) tableForT(t Type) (Table, bool) {
	switch t {
	case TInterface:
		return cc.interfaces, true

	case TController:
		return cc.controllers, true

	case TDoor:
		return cc.doors, true

	case TCard:
		return cc.cards, true

	case TGroup:
		return cc.groups, true

	case TEvent:
		return cc.events, true

	case TLog:
		return cc.logs, true

	case TUser:
		return cc.users, true

	default:
		return nil, false
	}
}

func (cc *db) tableFor(oid schema.OID) (Table, bool) {
	switch oid {
	case schema.InterfacesOID:
		return cc.interfaces, true

	case schema.ControllersOID:
		return cc.doors, true

	case schema.DoorsOID:
		return cc.doors, true

	case schema.CardsOID:
		return cc.cards, true

	case schema.GroupsOID:
		return cc.groups, true

	case schema.EventsOID:
		return cc.events, true

	case schema.LogsOID:
		return cc.logs, true

	case schema.UsersOID:
		return cc.users, true

	default:
		return nil, false
	}
}

// TODO Remove, pending migration to real Go generics
func TypeOf(v any) Type {
	t := fmt.Sprintf("%T", v)
	switch t {
	case "catalog.CatalogInterface":
		return TInterface

	case "catalog.CatalogController":
		return TController

	case "catalog.CatalogDoor":
		return TDoor

	case "catalog.CatalogCard":
		return TCard

	case "catalog.CatalogGroup":
		return TGroup

	case "catalog.CatalogEvent":
		return TEvent

	case "catalog.CatalogLogEntry":
		return TLog

	case "catalog.CatalogUser":
		return TUser

	default:
		return TUnknown
	}
}

// func tableForT[T ctypes.CatalogType](v T) Table {
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

// TODO REMOVE WHEN MIGRATION TO Go GENERICS IS DONE
type Type int

const (
	TUnknown Type = iota
	TInterface
	TController
	TDoor
	TCard
	TGroup
	TEvent
	TLog
	TUser
)

func (t Type) String() string {
	return []string{
		"unknown",
		"interface",
		"controller",
		"door",
		"card",
		"group",
		"event",
		"log",
		"user",
	}[t]
}
