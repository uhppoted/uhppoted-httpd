package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestEventsNewOID(t *testing.T) {
	tt := events{
		base: schema.EventsOID,
		m: map[schema.OID]*event{
			"0.6.1":  &event{},
			"0.6.2":  &event{},
			"0.6.10": &event{},
		},
		last: 123,
	}

	expected := events{
		base: schema.EventsOID,
		m: map[schema.OID]*event{
			"0.6.1":   &event{},
			"0.6.2":   &event{},
			"0.6.10":  &event{},
			"0.6.124": &event{},
		},
		last: 124,
	}

	oid := tt.New(struct{}{})

	if oid != "0.6.124" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.6.124", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to events\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestEventsPut(t *testing.T) {
	tt := events{
		base: schema.EventsOID,
		m: map[schema.OID]*event{
			"0.6.1":  &event{},
			"0.6.2":  &event{},
			"0.6.10": &event{},
		},
		last: 123,
	}

	expected := events{
		base: schema.EventsOID,
		m: map[schema.OID]*event{
			"0.6.1":   &event{},
			"0.6.2":   &event{},
			"0.6.10":  &event{},
			"0.6.124": &event{},
		},
		last: 124,
	}

	tt.Put("0.6.124", struct{}{})

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to events\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestEventsDelete(t *testing.T) {
	tt := events{
		m: map[schema.OID]*event{
			"0.6.1":  &event{},
			"0.6.2":  &event{},
			"0.6.10": &event{},
		},
	}

	expected := events{
		m: map[schema.OID]*event{
			"0.6.1": &event{},
			"0.6.2": &event{
				deleted: true,
			},
			"0.6.10": &event{},
		},
	}

	tt.Delete("0.6.2")

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("'delete' failed\n   expected:%v\n   got:     %v", expected, tt)

		for k, v := range tt.m {
			t.Errorf(">>> %v %v", k, v)
		}
	}
}
