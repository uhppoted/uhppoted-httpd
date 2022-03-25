package catalog

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestJoin(t *testing.T) {
	p := []schema.Object{
		schema.Object{OID: "0.1.1", Value: "A"},
		schema.Object{OID: "0.1.2", Value: "B"},
		schema.Object{OID: "0.1.3", Value: "C"},
	}

	q := []schema.Object{
		schema.Object{OID: "0.2.1", Value: "X"},
		schema.Object{OID: "0.3.2", Value: "Y"},
		schema.Object{OID: "0.3.3", Value: "Z"},
	}

	expected := []schema.Object{
		schema.Object{OID: "0.1.1", Value: "A"},
		schema.Object{OID: "0.1.2", Value: "B"},
		schema.Object{OID: "0.1.3", Value: "C"},
		schema.Object{OID: "0.2.1", Value: "X"},
		schema.Object{OID: "0.3.2", Value: "Y"},
		schema.Object{OID: "0.3.3", Value: "Z"},
	}

	Join(&p, q...)

	if !reflect.DeepEqual(p, expected) {
		t.Errorf("Object lists not joined correctly\n   expected:%v\n   got:     %v", expected, p)
	}
}

// func TestNewController(t *testing.T) {
// 	Init(stub{})
//
// 	type controller struct {
// 		CatalogController
// 	}
//
// 	catalog.Clear()
//
// 	p := controller{
// 		CatalogController: CatalogController{
// 			DeviceID: 1234,
// 		},
// 	}
//
// 	q := controller{
// 		CatalogController: CatalogController{
// 			DeviceID: 5678,
// 		},
// 	}
//
// 	r := controller{
// 		CatalogController: CatalogController{
// 			DeviceID: 1234,
// 		},
// 	}
//
// 	if oid := NewT(p.CatalogController); oid != "0.2.1" {
// 		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.1", oid)
// 	}
//
// 	if oid := NewT(q.CatalogController); oid != "0.2.2" {
// 		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.2", oid)
// 	}
//
// 	if oid := NewT(r.CatalogController); oid != "0.2.1" {
// 		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.1", oid)
// 	}
// }

// func TestNewDoor(t *testing.T) {
// 	type door struct {
// 		ctypes.CatalogDoor
// 	}
//
// 	catalog.Clear()
//
// 	oid := NewT(door{}.CatalogDoor)
//
// 	if oid != "0.3.1" {
// 		t.Errorf("Incorrect door OID - expected:%v, got:%v", "0.3.1", oid)
// 	}
// }

//func TestNewCard(t *testing.T) {
//	type card struct {
//		ctypes.CatalogCard
//	}
//
//	catalog.Clear()
//
//	oid := NewT(card{}.CatalogCard)
//
//	if oid != "0.4.1" {
//		t.Errorf("Incorrect card OID - expected:%v, got:%v", "0.4.1", oid)
//	}
//}

//func TestNewGroup(t *testing.T) {
//	type group struct {
//		ctypes.CatalogGroup
//	}
//
//	catalog.Clear()
//
//	oid := NewT(group{}.CatalogGroup)
//
//	if oid != "0.5.1" {
//		t.Errorf("Incorrect group OID - expected:%v, got:%v", "0.5.1", oid)
//	}
//}

//func TestNewEvent(t *testing.T) {
//	type event struct {
//		ctypes.CatalogEvent
//	}
//
//	catalog.Clear()
//
//	oid := NewT(event{}.CatalogEvent)
//
//	if oid != "0.6.1" {
//		t.Errorf("Incorrect event OID - expected:%v, got:%v", "0.6.1", oid)
//	}
//}

//func TestNewLogEntry(t *testing.T) {
//	type logentry struct {
//		ctypes.CatalogLogEntry
//	}
//
//	catalog.Clear()
//
//	oid := NewT(logentry{}.CatalogLogEntry)
//
//	if oid != "0.7.1" {
//		t.Errorf("Incorrect log entry OID - expected:%v, got:%v", "0.7.1", oid)
//	}
//}

//func TestNewUser(t *testing.T) {
//	type user struct {
//		ctypes.CatalogUser
//	}
//
//	catalog.Clear()
//
//	oid := NewT(user{}.CatalogUser)
//
//	if oid != "0.8.1" {
//		t.Errorf("Incorrect user OID - expected:%v, got:%v", "0.8.1", oid)
//	}
//}
