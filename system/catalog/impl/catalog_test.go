package memdb

import (
	"reflect"
	"sort"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestNewInterface(t *testing.T) {
	type lan struct {
		catalog.CatalogInterface
	}

	Catalog().Clear()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected 'panic' - currently only a single static LAN interface is supported")
		}
	}()

	Catalog().NewT(lan{}.CatalogInterface)
}

func TestNewController(t *testing.T) {
	type controller struct {
		catalog.CatalogController
	}

	Catalog().Clear()

	p := controller{
		CatalogController: catalog.CatalogController{
			DeviceID: 1234,
		},
	}

	q := controller{
		CatalogController: catalog.CatalogController{
			DeviceID: 5678,
		},
	}

	r := controller{
		CatalogController: catalog.CatalogController{
			DeviceID: 1234,
		},
	}

	if oid := Catalog().NewT(p.CatalogController); oid != "0.2.1" {
		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.1", oid)
	}

	if oid := Catalog().NewT(q.CatalogController); oid != "0.2.2" {
		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.2", oid)
	}

	if oid := Catalog().NewT(r.CatalogController); oid != "0.2.1" {
		t.Errorf("Incorrect controller OID - expected:%v, got:%v", "0.2.1", oid)
	}
}

func TestNewDoor(t *testing.T) {
	type door struct {
		catalog.CatalogDoor
	}

	cc := db{
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

	expected := db{
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

	oid := cc.NewT(door{}.CatalogDoor)

	if oid != "0.3.101" {
		t.Errorf("Incorrect OID - expected:%v, got:%v", "0.3.101", oid)
	}

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestNewCard(t *testing.T) {
	type card struct {
		catalog.CatalogCard
	}

	Catalog().Clear()

	oid := Catalog().NewT(card{}.CatalogCard)

	if oid != "0.4.1" {
		t.Errorf("Incorrect card OID - expected:%v, got:%v", "0.4.1", oid)
	}
}

func TestNewGroup(t *testing.T) {
	type group struct {
		catalog.CatalogGroup
	}

	Catalog().Clear()

	oid := Catalog().NewT(group{}.CatalogGroup)

	if oid != "0.5.1" {
		t.Errorf("Incorrect group OID - expected:%v, got:%v", "0.5.1", oid)
	}
}

func TestNewEvent(t *testing.T) {
	type event struct {
		catalog.CatalogEvent
	}

	cc := db{
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
		oid := cc.NewT(event{}.CatalogEvent)

		if oid != expected {
			t.Errorf("Invalid event OID - expected:%v, got:%v", expected, oid)
		}
	}
}

func TestNewLogEntry(t *testing.T) {
	type logentry struct {
		catalog.CatalogLogEntry
	}

	Catalog().Clear()

	oid := Catalog().NewT(logentry{}.CatalogLogEntry)

	if oid != "0.7.1" {
		t.Errorf("Incorrect log entry OID - expected:%v, got:%v", "0.7.1", oid)
	}
}

func TestNewUser(t *testing.T) {
	type user struct {
		catalog.CatalogUser
	}

	Catalog().Clear()

	oid := Catalog().NewT(user{}.CatalogUser)

	if oid != "0.8.1" {
		t.Errorf("Incorrect user OID - expected:%v, got:%v", "0.8.1", oid)
	}
}

func TestListT(t *testing.T) {
	cc := db{
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

	list := cc.ListT(schema.DoorsOID)

	sort.Slice(list, func(i, j int) bool { return string(list[i]) < string(list[j]) })

	if !reflect.DeepEqual(&list, &expected) {
		t.Errorf("Incorrect list of doors:\n   expected:%v\n   got:     %v", &expected, &list)
	}
}

func TestHasT(t *testing.T) {
	type group struct {
		catalog.CatalogGroup
	}

	cc := db{
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
		if has := cc.HasT(group{}.CatalogGroup, k); has != v {
			t.Errorf("HasT returned incorrect result for '%v' - expected:%v\n, got:%v", k, v, has)
		}
	}
}

func TestDeleteT(t *testing.T) {
	type door struct {
		catalog.CatalogDoor
	}

	cc := db{
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

	expected := db{
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

	cc.DeleteT(door{}.CatalogDoor, "0.3.3")

	if !reflect.DeepEqual(&cc, &expected) {
		t.Errorf("Catalog not updated:\n   expected:%v\n   got:     %v", &expected, &cc)
	}
}

func TestClear(t *testing.T) {
	cc := db{
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

	expected := db{
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
