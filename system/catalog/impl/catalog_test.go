package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

func TestNewOID(t *testing.T) {
	cc := catalog{
		doors: table{
			base: schema.DoorsOID,
			m: map[schema.OID]struct{}{
				"0.3.1":   struct{}{},
				"0.3.2":   struct{}{},
				"0.3.100": struct{}{},
			},
		},

		controllers: map[schema.OID]controller{},
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
			m: map[schema.OID]struct{}{
				"0.3.1":   struct{}{},
				"0.3.2":   struct{}{},
				"0.3.3":   struct{}{},
				"0.3.100": struct{}{},
			},
		},

		controllers: map[schema.OID]controller{},
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
			m: map[schema.OID]struct{}{
				"0.3.1":   struct{}{},
				"0.3.2":   struct{}{},
				"0.3.100": struct{}{},
			},
		},

		controllers: map[schema.OID]controller{},
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
			m: map[schema.OID]struct{}{
				"0.3.1":   struct{}{},
				"0.3.2":   struct{}{},
				"0.3.3":   struct{}{},
				"0.3.100": struct{}{},
			},
		},

		controllers: map[schema.OID]controller{},
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
			m:    map[schema.OID]struct{}{},
		},

		controllers: map[schema.OID]controller{},
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
