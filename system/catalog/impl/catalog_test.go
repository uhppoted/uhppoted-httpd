package memdb

import (
	"reflect"
	"sort"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

func TestNewOID(t *testing.T) {
	cc := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	expected := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.3":   entry{},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	oid := cc.NewT(ctypes.TDoor, struct{}{})

	if oid != "0.3.3" {
		t.Errorf("Incorrect OID - expected:%v, got:%v", "0.3.3", oid)
	}

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestNewDoor(t *testing.T) {
	cc := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	expected := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.3":   entry{},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	oid := cc.NewT(ctypes.TDoor, nil)

	if oid != "0.3.3" {
		t.Errorf("Incorrect OID - expected:%v, got:%v", "0.3.3", oid)
	}

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestNewEvent(t *testing.T) {
	cc := catalog{
		events: table{
			base: schema.EventsOID,
			m:    map[schema.OID]entry{},
		},

		controllers: controllers{},
		interfaces:  table{},
		doors:       table{},
		cards:       table{},
		groups:      table{},
		logs:        table{},
		users:       table{},
	}

	tests := []schema.OID{
		schema.OID("0.6.1"),
		schema.OID("0.6.2"),
		schema.OID("0.6.3"),
	}

	for _, expected := range tests {
		oid := cc.NewT(ctypes.TEvent, nil)

		if oid != expected {
			t.Errorf("Invalid event OID - expected:%v, got:%v", expected, oid)
		}
	}
}

func TestListT(t *testing.T) {
	cc := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.3":   entry{deleted: true},
				"0.3.100": entry{},
				"0.3.200": entry{},
			},
		},
	}

	expected := []schema.OID{
		"0.3.1",
		"0.3.100",
		"0.3.2",
		"0.3.200",
	}

	list := cc.ListT(ctypes.TDoor)

	sort.Slice(list, func(i, j int) bool { return string(list[i]) < string(list[j]) })

	if !reflect.DeepEqual(&list, &expected) {
		t.Errorf("Incorrect list of doors:\n   expected:%v\n   got:     %v", &expected, &list)
	}
}

func TestHasT(t *testing.T) {
	cc := catalog{
		groups: table{
			base: schema.GroupsOID,
			m: map[schema.OID]entry{
				"0.5.1":   entry{},
				"0.5.2":   entry{},
				"0.5.3":   entry{deleted: true},
				"0.5.100": entry{},
				"0.5.200": entry{},
			},
		},
	}

	tests := map[schema.OID]bool{
		"0.5.1":   true,
		"0.5.2":   true,
		"0.5.3":   false,
		"0.5.100": true,
		"0.5.200": true,
		"0.5.5":   false,
	}

	for k, v := range tests {
		if has := cc.HasT(ctypes.TGroup, k); has != v {
			t.Errorf("HasT returned incorrect result for '%v' - expected:%v\n, got:%v", k, v, has)
		}
	}
}

func TestDeleteT(t *testing.T) {
	cc := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.3":   entry{},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	expected := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]entry{
				"0.3.1":   entry{},
				"0.3.2":   entry{},
				"0.3.3":   entry{deleted: true},
				"0.3.100": entry{},
			},
		},

		controllers: controllers{},
		interfaces:  table{},
		cards:       table{},
		groups:      table{},
		events:      table{},
		logs:        table{},
		users:       table{},
	}

	cc.DeleteT(ctypes.TDoor, "0.3.3")

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}
