package history

import (
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
)

var entries = []Entry{
	Entry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Value:     "Alpha1",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Value:     "Alpha2",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Value:     "Alpha3",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Value:     "Alpha4",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Value:     "Barney1",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Value:     "Barney2",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Value:     "Barney3",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Value:     "Barney4",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Value:     "Cupboard1",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Value:     "Cupboard2",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Value:     "Cupboard3",
	},

	Entry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Value:     "Cupboard4",
	},
}

func TestLookupDefaultControllerName(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := ""

	history := NewHistory()
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupControllerWithoutRelevantHistory(t *testing.T) {
	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 303986753})
	catalog.PutV("0.2.1", schema.ControllerName, "Beta")

	expected := "Beta"

	history := NewHistory(entries...)
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	controller := history.LookupController(timestamp, 303986753)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupController(t *testing.T) {
	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
	catalog.PutV("0.2.1", schema.ControllerName, "Alpha")

	expected := "Alpha3"

	history := NewHistory(entries...)
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupPrehistoricController(t *testing.T) {
	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
	catalog.PutV("0.2.1", schema.ControllerName, "Alpha")

	expected := ""

	history := NewHistory(entries...)
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupFashionablyNewController(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := "Alpha4"

	history := NewHistory(entries...)
	timestamp := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupFashionablyNewController2(t *testing.T) {
	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
	catalog.PutV("0.2.1", schema.ControllerName, "Alpha")

	expected := "Alpha"

	history := NewHistory(entries...)
	timestamp := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupDefaultCardName(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := ""

	history := NewHistory()
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	name := history.LookupCard(timestamp, 8165538)

	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupCardName(t *testing.T) {
	card := cards.Card{
		CatalogCard: catalog.CatalogCard{
			OID: "0.4.1",
		},
	}

	catalog.Init(memdb.NewCatalog())
	catalog.PutT(card.CatalogCard)
	catalog.PutV("0.4.1", schema.CardNumber, uint32(8165538))
	catalog.PutV("0.4.1", schema.CardName, "FredF")

	expected := "FredF"

	history := NewHistory()
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	name := history.LookupCard(timestamp, 8165538)

	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalCardName(t *testing.T) {
	card := cards.Card{
		CatalogCard: catalog.CatalogCard{
			OID: "0.4.1",
		},
	}

	catalog.Init(memdb.NewCatalog())
	catalog.PutT(card.CatalogCard)
	catalog.PutV("0.4.1", schema.CardNumber, uint32(8165538))
	catalog.PutV("0.4.1", schema.CardName, "FredF")

	expected := "Barney3"

	history := NewHistory(entries...)
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	name := history.LookupCard(timestamp, 8165538)

	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDefaultDoorName(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := ""

	history := NewHistory()
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	door := history.LookupDoor(timestamp, 405419896, 3)

	if door != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, door)
	}
}

func TestLookupDoorName(t *testing.T) {
	door := doors.Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.1",
		},
	}

	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
	catalog.PutV("0.2.1", schema.ControllerName, "Alpha")
	catalog.PutV("0.2.1", schema.ControllerDeviceID, 405419896)
	catalog.PutV("0.2.1", schema.ControllerDoor3, schema.OID("0.3.1"))
	catalog.PutT(door.CatalogDoor)
	catalog.PutV("0.3.1", schema.DoorName, "Gringotts")

	expected := "Gringotts"

	history := NewHistory()
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	name := history.LookupDoor(timestamp, 405419896, 3)

	if name != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalDoorName(t *testing.T) {
	door := doors.Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.1",
		},
	}

	catalog.Init(memdb.NewCatalog())
	catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
	catalog.PutV("0.2.1", schema.ControllerName, "Alpha")
	catalog.PutV("0.2.1", schema.ControllerDeviceID, 405419896)
	catalog.PutV("0.2.1", schema.ControllerDoor3, schema.OID("0.3.1"))
	catalog.PutT(door.CatalogDoor)
	catalog.PutV("0.3.1", schema.DoorName, "Gringotts")

	expected := "Cupboard3"

	history := NewHistory(entries...)
	timestamp := time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)
	name := history.LookupDoor(timestamp, 405419896, 3)

	if name != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
	}
}
