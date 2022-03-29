package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestTableNewOID(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m: map[schema.OID]*record{
			"0.3.1":  &record{},
			"0.3.2":  &record{},
			"0.3.10": &record{},
		},
		last: 123,
	}

	expected := table{
		base: schema.DoorsOID,
		m: map[schema.OID]*record{
			"0.3.1":   &record{},
			"0.3.2":   &record{},
			"0.3.10":  &record{},
			"0.3.124": &record{},
		},
		last: 124,
	}

	oid := tt.New(struct{}{})

	if oid != "0.3.124" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.3.124", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTablePut(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m: map[schema.OID]*record{
			"0.3.1":  &record{},
			"0.3.2":  &record{},
			"0.3.10": &record{},
		},
		last: 123,
	}

	expected := table{
		base: schema.DoorsOID,
		m: map[schema.OID]*record{
			"0.3.1":   &record{},
			"0.3.2":   &record{},
			"0.3.10":  &record{},
			"0.3.124": &record{},
		},
		last: 124,
	}

	tt.Put("0.3.124", struct{}{})

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTableDelete(t *testing.T) {
	tt := table{
		m: map[schema.OID]*record{
			"0.3.1":  &record{},
			"0.3.2":  &record{},
			"0.3.10": &record{},
		},
	}

	expected := table{
		m: map[schema.OID]*record{
			"0.3.1": &record{},
			"0.3.2": &record{
				deleted: true,
			},
			"0.3.10": &record{},
		},
	}

	tt.Delete("0.3.2")

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("'delete' failed\n   expected:%v\n   got:     %v", expected, tt)

		for k, v := range tt.m {
			t.Errorf(">>> %v %v", k, v)
		}
	}
}
