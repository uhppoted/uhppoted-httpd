package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestTableNewTWithCompacting(t *testing.T) {
	tt := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":  &entry{},
			"0.3.2":  &entry{},
			"0.3.10": &entry{},
		},
		last:  10,
		limit: 32,
	}

	expected := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":  &entry{},
			"0.3.2":  &entry{},
			"0.3.3":  &entry{},
			"0.3.10": &entry{},
		},
		last:  10,
		limit: 32,
	}
	oid := newOID(tt, struct{}{})

	if oid != "0.3.3" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.3.3", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTableNewTWithoutCompacting(t *testing.T) {
	tt := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":  &entry{},
			"0.3.2":  &entry{},
			"0.3.10": &entry{},
		},
		last:  123,
		limit: -1,
	}

	expected := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":   &entry{},
			"0.3.2":   &entry{},
			"0.3.10":  &entry{},
			"0.3.124": &entry{},
		},
		last:  123,
		limit: -1,
	}

	oid := newOID(tt, struct{}{})

	if oid != "0.3.124" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.3.124", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTablePut(t *testing.T) {
	tt := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":  &entry{},
			"0.3.2":  &entry{},
			"0.3.10": &entry{},
		},
		last:  123,
		limit: -1,
	}

	expected := table[*entry]{
		base: schema.DoorsOID,
		m: map[schema.OID]*entry{
			"0.3.1":   &entry{},
			"0.3.2":   &entry{},
			"0.3.10":  &entry{},
			"0.3.124": &entry{},
		},
		last:  123,
		limit: -1,
	}

	put(tt, "0.3.124", struct{}{})

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTableDelete(t *testing.T) {
	tt := table[*entry]{
		m: map[schema.OID]*entry{
			"0.3.1":  &entry{},
			"0.3.2":  &entry{},
			"0.3.10": &entry{},
		},
	}

	expected := table[*entry]{
		m: map[schema.OID]*entry{
			"0.3.1": &entry{},
			"0.3.2": &entry{
				deleted: true,
			},
			"0.3.10": &entry{},
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

func TestTableNewController(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last:  10,
		limit: 32,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
			"0.2.11": &controller{ID: 1234},
		},
		last:  11,
		limit: 32,
	}

	oid := tt.New(1234)

	if oid != "0.2.11" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.2.11", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTablePutController(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last:  123,
		limit: -1,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":   &controller{ID: 1001},
			"0.2.2":   &controller{ID: 1002},
			"0.2.10":  &controller{ID: 1010},
			"0.2.124": &controller{ID: 1234},
		},
		last:  124,
		limit: -1,
	}

	tt.Put("0.2.124", 1234)

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to controllers\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestTableDeleteController(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last:  123,
		limit: -1,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002, deleted: true},
			"0.2.10": &controller{ID: 1010},
		},
		last:  123,
		limit: -1,
	}

	tt.Delete("0.2.2")

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("'delete' failed\n   expected:%v\n   got:     %v", expected, tt)

		for k, v := range tt.m {
			t.Errorf(">>> %v %v", k, v)
		}
	}
}
