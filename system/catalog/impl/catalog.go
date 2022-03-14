package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type catalog struct {
	interfaces  table
	controllers map[schema.OID]controller
	doors       table
	cards       table
	groups      table
	events      table
	logs        table
	users       table
	guard       sync.Mutex
}

type table struct {
	base schema.OID
	m    map[schema.OID]struct{}
	last uint32
}

type controller struct {
	ID      uint32
	deleted bool
}

var db = catalog{
	controllers: map[schema.OID]controller{},

	interfaces: table{
		base: schema.InterfacesOID,
		m:    map[schema.OID]struct{}{},
	},
	doors: table{
		base: schema.DoorsOID,
		m:    map[schema.OID]struct{}{},
	},
	cards: table{
		base: schema.CardsOID,
		m:    map[schema.OID]struct{}{},
	},
	groups: table{
		base: schema.GroupsOID,
		m:    map[schema.OID]struct{}{},
	},
	events: table{
		base: schema.EventsOID,
		m:    map[schema.OID]struct{}{},
	},
	logs: table{
		base: schema.LogsOID,
		m:    map[schema.OID]struct{}{},
	},
	users: table{
		base: schema.UsersOID,
		m:    map[schema.OID]struct{}{},
	},
}

func Catalog() *catalog {
	return &db
}

func (cc *catalog) Clear() {
	cc.controllers = map[schema.OID]controller{}

	cc.interfaces.m = map[schema.OID]struct{}{}
	cc.doors.m = map[schema.OID]struct{}{}
	cc.cards.m = map[schema.OID]struct{}{}
	cc.groups.m = map[schema.OID]struct{}{}
	cc.events.m = map[schema.OID]struct{}{}
	cc.logs.m = map[schema.OID]struct{}{}

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
	// NTS: only support a single interface at this point in time
	if t == ctypes.TInterface {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	if t == ctypes.TController {
		return cc.newController(v.(uint32))
	}

	m, ok := cc.tableFor(t)
	if !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	}

	cc.guard.Lock()
	defer cc.guard.Unlock()

	return m.NewT(v)
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

	if m, ok := cc.tableFor(t); !ok {
		panic(fmt.Sprintf("Unsupported catalog type (%v)", t))
	} else {
		m.Put(oid, v)
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
	return cc.doors.m
}

func (cc *catalog) Groups() map[schema.OID]struct{} {
	return cc.groups.m
}

func (cc *catalog) HasGroup(oid schema.OID) bool {
	cc.guard.Lock()
	defer cc.guard.Unlock()

	_, ok := cc.groups.m[oid]

	return ok
}

func (cc *catalog) tableFor(t ctypes.Type) (table, bool) {
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
		return table{}, false
	}
}

func (t *table) NewT(v interface{}) schema.OID {
	suffix := t.last

loop:
	for {
		suffix += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", t.base, suffix))
		for v, _ := range t.m {
			if v == oid {
				continue loop
			}
		}

		t.m[oid] = struct{}{}
		t.last = suffix
		return oid
	}
}

func (t *table) Put(oid schema.OID, v interface{}) {
	if !oid.HasPrefix(t.base) {
		panic(fmt.Sprintf("PUT: illegal oid %v for base %v", oid, t.base))
	}

	suffix := strings.TrimPrefix(string(oid), string(t.base))

	match := regexp.MustCompile(`\.([0-9]+)`).FindStringSubmatch(suffix)
	if match == nil || len(match) != 2 {
		panic(fmt.Sprintf("PUT: invalid oid %v for base %v", oid, t.base))
	}

	index, err := strconv.ParseUint(match[1], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("PUT: out of range oid %v for base %v", oid, t.base))
	}

	t.m[oid] = struct{}{}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}
