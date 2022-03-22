package memdb

import (
	"reflect"
	"sort"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

func TestNewDoor(t *testing.T) {
	cc := catalog{
		doors: &table{
			base: schema.DoorsOID,
			m: map[schema.OID]*record{
				"0.3.1":   &record{},
				"0.3.2":   &record{},
				"0.3.100": &record{},
			},
			last: 100,
		},

		controllers: &controllers{},
		interfaces:  &table{},
		cards:       &table{},
		groups:      &table{},
		events:      &table{},
		logs:        &table{},
		users:       &table{},
	}

	expected := catalog{
		doors: &table{
			base: schema.DoorsOID,
			m: map[schema.OID]*record{
				"0.3.1":   &record{},
				"0.3.2":   &record{},
				"0.3.100": &record{},
				"0.3.101": &record{},
			},
			last: 101,
		},

		controllers: &controllers{},
		interfaces:  &table{},
		cards:       &table{},
		groups:      &table{},
		events:      &table{},
		logs:        &table{},
		users:       &table{},
	}

	oid := cc.NewT(ctypes.TDoor, nil)

	if oid != "0.3.101" {
		t.Errorf("Incorrect OID - expected:%v, got:%v", "0.3.101", oid)
	}

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestNewEvent(t *testing.T) {
	cc := catalog{
		events: &table{
			base: schema.EventsOID,
			m:    map[schema.OID]*record{},
		},

		controllers: &controllers{},
		interfaces:  &table{},
		doors:       &table{},
		cards:       &table{},
		groups:      &table{},
		logs:        &table{},
		users:       &table{},
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
		doors: &table{
			base: schema.DoorsOID,
			m: map[schema.OID]*record{
				"0.3.1":   &record{},
				"0.3.2":   &record{},
				"0.3.3":   &record{deleted: true},
				"0.3.100": &record{},
				"0.3.200": &record{},
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
		groups: &table{
			base: schema.GroupsOID,
			m: map[schema.OID]*record{
				"0.5.1":   &record{},
				"0.5.2":   &record{},
				"0.5.3":   &record{deleted: true},
				"0.5.100": &record{},
				"0.5.200": &record{},
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
		doors: &table{
			base: schema.DoorsOID,
			m: map[schema.OID]*record{
				"0.3.1":   &record{},
				"0.3.2":   &record{},
				"0.3.3":   &record{},
				"0.3.100": &record{},
			},
		},

		controllers: &controllers{},
		interfaces:  &table{},
		cards:       &table{},
		groups:      &table{},
		events:      &table{},
		logs:        &table{},
		users:       &table{},
	}

	expected := catalog{
		doors: &table{
			base: schema.DoorsOID,
			m: map[schema.OID]*record{
				"0.3.1":   &record{},
				"0.3.2":   &record{},
				"0.3.3":   &record{deleted: true},
				"0.3.100": &record{},
			},
		},

		controllers: &controllers{},
		interfaces:  &table{},
		cards:       &table{},
		groups:      &table{},
		events:      &table{},
		logs:        &table{},
		users:       &table{},
	}

	cc.DeleteT(ctypes.TDoor, "0.3.3")

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestClear(t *testing.T) {
	cc := catalog{
		interfaces: &table{
			base: schema.InterfacesOID,
			m:    map[schema.OID]*record{"0.1.1": &record{}},
			last: 101,
		},
		controllers: &controllers{
			base: schema.ControllersOID,
			m:    map[schema.OID]*controller{"0.2.1": &controller{}},
			last: 102,
		},
		doors: &table{
			base: schema.DoorsOID,
			m:    map[schema.OID]*record{"0.3.1": &record{}},
			last: 103,
		},
		cards: &table{
			base: schema.CardsOID,
			m:    map[schema.OID]*record{"0.4.1": &record{}},
			last: 104,
		},
		groups: &table{
			base: schema.GroupsOID,
			m:    map[schema.OID]*record{"0.5.1": &record{}},
			last: 105,
		},
		events: &table{
			base: schema.EventsOID,
			m:    map[schema.OID]*record{"0.6.1": &record{}},
			last: 106,
		},
		logs: &table{
			base: schema.LogsOID,
			m:    map[schema.OID]*record{"0.7.1": &record{}},
			last: 107,
		},
		users: &table{
			base: schema.UsersOID,
			m:    map[schema.OID]*record{"0.8.1": &record{}},
			last: 108,
		},
	}

	expected := catalog{
		interfaces: &table{
			base: schema.InterfacesOID,
			m:    map[schema.OID]*record{},
			last: 0},
		controllers: &controllers{
			base: schema.ControllersOID,
			m:    map[schema.OID]*controller{},
			last: 0},
		doors: &table{
			base: schema.DoorsOID,
			m:    map[schema.OID]*record{},
			last: 0},
		cards: &table{
			base: schema.CardsOID,
			m:    map[schema.OID]*record{},
			last: 0},
		groups: &table{
			base: schema.GroupsOID,
			m:    map[schema.OID]*record{},
			last: 0},
		events: &table{
			base: schema.EventsOID,
			m:    map[schema.OID]*record{},
			last: 0},
		logs: &table{
			base: schema.LogsOID,
			m:    map[schema.OID]*record{},
			last: 0},
		users: &table{
			base: schema.UsersOID,
			m:    map[schema.OID]*record{},
			last: 0},
	}

	cc.Clear()

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%#v\n   got:     %#v", &expected, &cc)
	}
}
