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

func NewCatalog() *db {
	return &db{
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

func (cc *db) FindController(v catalog.CatalogController) schema.OID {
	cc.guard.RLock()
	defer cc.guard.RUnlock()

	if t, ok := cc.controllers.(*controllers); ok {
		return t.Find(v)
	}

	return ""
}

// NTS: There really doesn't really seem to be a way to do this with Go generics
//
// Ref. https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#methods-may-not-take-additional-type-arguments
type tableType interface {
	TypeOf() catalog.Type
}

func tableForT(cc *db, v any) Table {
	if t, ok := v.(tableType); ok {
		switch t.TypeOf() {

		case catalog.TInterface:
			return cc.interfaces

		case catalog.TController:
			return cc.controllers

		case catalog.TDoor:
			return cc.doors

		case catalog.TCard:
			return cc.cards

		case catalog.TGroup:
			return cc.groups

		case catalog.TEvent:
			return cc.events

		case catalog.TLogEntry:
			return cc.logs

		case catalog.TUser:
			return cc.users
		}
	}

	return nil
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
