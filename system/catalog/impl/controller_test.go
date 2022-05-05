package memdb

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestControllerNew(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last: 10,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
			"0.2.11": &controller{ID: 1234},
		},
		last: 11,
	}

	oid := tt.New(catalog.CatalogController{
		DeviceID: 1234,
	})

	if oid != "0.2.11" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", "0.2.11", oid)
	}

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("New OID not added to table\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestControllerPut(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last: 123,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":   &controller{ID: 1001},
			"0.2.2":   &controller{ID: 1002},
			"0.2.10":  &controller{ID: 1010},
			"0.2.124": &controller{ID: 1234},
		},
		last: 124,
	}

	tt.Put("0.2.124", catalog.CatalogController{
		DeviceID: 1234,
	})

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("OID not added to controllers\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestControllerDelete(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002},
			"0.2.10": &controller{ID: 1010},
		},
		last: 123,
	}

	expected := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1":  &controller{ID: 1001},
			"0.2.2":  &controller{ID: 1002, deleted: true},
			"0.2.10": &controller{ID: 1010},
		},
		last: 123,
	}

	tt.Delete("0.2.2")

	if !reflect.DeepEqual(tt, expected) {
		t.Errorf("'delete' failed\n   expected:%v\n   got:     %v", expected, tt)
	}
}

func TestControllersFind(t *testing.T) {
	tt := controllers{
		base: schema.ControllersOID,
		m: map[schema.OID]*controller{
			"0.2.1": &controller{
				ID: 1234678,
			},
			"0.2.7": &controller{
				ID: 23456789,
			},
			"0.2.89": &controller{
				ID: 34567890,
			},
		},
		last: 100,
	}

	if oid := tt.Find(catalog.CatalogController{DeviceID: 23456789}); oid != "0.2.7" {
		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.7", oid)
	}

	if oid := tt.Find(catalog.CatalogController{DeviceID: 45678901}); oid != "" {
		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "", oid)
	}
}
