package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestEventsNewOID(t *testing.T) {
	tt := events{
		base: schema.EventsOID,
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{},
			eventKey{405419896, 2}:  &event{},
			eventKey{405419896, 10}: &event{},
		},
		last: 123,
	}

	expected := events{
		base: schema.EventsOID,
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{},
			eventKey{405419896, 2}:  &event{},
			eventKey{405419896, 10}: &event{},
			eventKey{405419896, 999}: &event{
				OID:      "0.6.124",
				deviceID: 405419896,
				index:    999,
			},
		},
		last: 124,
	}

	oid := tt.New(catalog.CatalogEvent{
		DeviceID: 405419896,
		Index:    999,
	})

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
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{},
			eventKey{405419896, 2}:  &event{},
			eventKey{405419896, 10}: &event{},
		},
		last: 123,
	}

	expected := events{
		base: schema.EventsOID,
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{},
			eventKey{405419896, 2}:  &event{},
			eventKey{405419896, 10}: &event{},
			eventKey{405419896, 999}: &event{
				OID:      "0.6.124",
				deviceID: 405419896,
				index:    999,
			},
		},
		last: 124,
	}

	tt.Put("0.6.124", catalog.CatalogEvent{
		OID:      "0.6.124",
		DeviceID: 405419896,
		Index:    999,
	})

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to events\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestEventsDelete(t *testing.T) {
	tt := events{
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{OID: "0.6.1"},
			eventKey{405419896, 2}:  &event{OID: "0.6.2"},
			eventKey{405419896, 10}: &event{OID: "0.6.10"},
		},
	}

	expected := events{
		m: map[eventKey]*event{
			eventKey{405419896, 1}:  &event{OID: "0.6.1"},
			eventKey{405419896, 2}:  &event{OID: "0.6.2", deleted: true},
			eventKey{405419896, 10}: &event{OID: "0.6.10"},
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
