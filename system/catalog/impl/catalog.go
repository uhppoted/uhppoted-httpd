package memdb

import (
	"fmt"
	"sync"

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
	tt := tableForT(cc, v)
	if tt == nil {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	// NTS: only supports a 'known' interface at this point in time
	if tt == cc.interfaces {
		panic(fmt.Sprintf("Unsupported catalog type (%T)", v))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	return tt.New(v)
}

func (cc *db) PutT(v any, oid schema.OID) {
	tt := tableForT(cc, v)
	if tt == nil {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	tt.Put(oid, v)
}

func (cc *db) DeleteT(v any, oid schema.OID) {
	tt := tableForT(cc, v)
	if tt == nil {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	tt.Delete(oid)
}

func (cc *db) ListT(oid schema.OID) []schema.OID {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	if tt := tableFor(cc, oid); tt != nil {
		return tt.List()
	}

	return []schema.OID{}
}

func (cc *db) HasT(v any, oid schema.OID) bool {
	tt := tableForT(cc, v)
	if tt == nil {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	}

	cc.guard.RLock()
	defer cc.guard.RUnlock()

	return tt.Has(v, oid)
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

// TODO Ewwww - but there doesn't seem to be an elegant way with the
//              current state of Go generics :-(
func tableForT(cc *db, v any) Table {
	t := fmt.Sprintf("%T", v)

	switch t {
	case "catalog.CatalogInterface":
		return cc.interfaces

	case "catalog.CatalogController":
		return cc.controllers

	case "catalog.CatalogDoor":
		return cc.doors

	case "catalog.CatalogCard":
		return cc.cards

	case "catalog.CatalogGroup":
		return cc.groups

	case "catalog.CatalogEvent":
		return cc.events

	case "catalog.CatalogLogEntry":
		return cc.logs

	case "catalog.CatalogUser":
		return cc.users

	default:
		return nil
	}
}

func tableFor(cc *db, oid schema.OID) Table {
	switch oid {
	case schema.InterfacesOID:
		return cc.interfaces

	case schema.ControllersOID:
		return cc.doors

	case schema.DoorsOID:
		return cc.doors

	case schema.CardsOID:
		return cc.cards

	case schema.GroupsOID:
		return cc.groups

	case schema.EventsOID:
		return cc.events

	case schema.LogsOID:
		return cc.logs

	case schema.UsersOID:
		return cc.users

	default:
		return nil
	}
}

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
