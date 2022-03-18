package memdb

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestTabletNewTWithCompacting(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m: map[schema.OID]record{
			"0.3.1":  record{},
			"0.3.2":  record{},
			"0.3.10": record{},
		},
		last:  10,
		limit: 32,
	}

	expected := schema.OID("0.3.3")
	oid := tt.New(struct{}{})

	if oid != expected {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", expected, oid)
	}
}

func TestTabletNewTWithoutCompacting(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m: map[schema.OID]record{
			"0.3.1":  record{},
			"0.3.2":  record{},
			"0.3.10": record{},
		},
		last:  123,
		limit: -1,
	}

	expected := schema.OID("0.3.124")
	oid := tt.New(struct{}{})

	if oid != expected {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", expected, oid)
	}
}
