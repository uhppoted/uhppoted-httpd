package memdb

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestTabletNewTWithCompacting(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m: map[schema.OID]entry{
			"0.3.1":  entry{},
			"0.3.2":  entry{},
			"0.3.10": entry{},
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
		m: map[schema.OID]entry{
			"0.3.1":  entry{},
			"0.3.2":  entry{},
			"0.3.10": entry{},
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

// func TestGenericTable(t *testing.T) {
// 	tt := Table[*entry]{
// 		m: map[schema.OID]*entry{
// 			"0.3.1":  &entry{},
// 			"0.3.2":  &entry{},
// 			"0.3.10": &entry{},
// 		},
// 	}
// 	expected := Table[*entry]{
// 		m: map[schema.OID]*entry{
// 			"0.3.1": &entry{},
// 			"0.3.2": &entry{
// 				deleted: true,
// 			},
// 			"0.3.10": &entry{},
// 		},
// 	}
//
// 	tt.Delete("0.3.2")
//
// 	if !reflect.DeepEqual(tt, expected) {
// 		t.Errorf("'delete' failed\n   expected:%v\n   got:     %v", expected, tt)
//
// 		for k, v := range tt.m {
// 			t.Errorf(">>> %v %v", k, v)
// 		}
// 	}
// }
