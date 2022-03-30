package system

import (
	"crypto/sha1"
	"testing"
	"time"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/logs"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

var event = uhppoted.Event{
	DeviceID:   405419896,
	Index:      17,
	Type:       1,
	Granted:    true,
	Door:       3,
	Direction:  1,
	CardNumber: 8165538,
	Timestamp:  types.DateTime(time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)),
	Reason:     1,
}

var history = map[[20]byte]logs.LogEntry{}

func init() {
	hash := func(s string) [20]byte {
		return sha1.Sum([]byte(s))
	}

	history[hash("CONTROLLER.1")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha1",
		After:     "Alpha2",
	}

	history[hash("CONTROLLER.2")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha3",
		After:     "Alpha4",
	}

	history[hash("CONTROLLER.3")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha5",
		After:     "Alpha6",
	}

	history[hash("CONTROLLER.4")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha7",
		After:     "Alpha8",
	}

	history[hash("CARD.1")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Before:    "Card1",
		After:     "Card2",
	}

	history[hash("CARD.2")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Before:    "Card3",
		After:     "Card4",
	}

	history[hash("CARD.3")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Before:    "Card5",
		After:     "Card6",
	}

	history[hash("CARD.4")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "card",
		ItemID:    "8165538",
		Field:     "name",
		Before:    "Barney",
		After:     "Card8",
	}

	history[hash("DOOR.1")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Before:    "Door1",
		After:     "Door2",
	}

	history[hash("DOOR.2")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Before:    "Door3",
		After:     "Door4",
	}

	history[hash("DOOR.3")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Before:    "Door5",
		After:     "Door6",
	}

	history[hash("DOOR.4")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "door",
		ItemID:    "405419896:3",
		Field:     "name",
		Before:    "Cupboard",
		After:     "Door8",
	}
}

func TestLookupDefaultDeviceName(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	sys.logs.Logs = logs.NewLogs()

	expected := ""

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDeviceNameWithoutRelevantLogs(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	h := []logs.LogEntry{}
	for _, v := range history {
		if v.Timestamp.Before(time.Time(event.Timestamp)) {
			h = append(h, v)
		}
	}

	sys.logs.Logs = logs.NewLogs(h...)

	oid := schema.OID("0.2.1")
	expected := "Alpha"

	catalog.PutT(catalog.CatalogController{DeviceID: 405419896}, oid)
	catalog.PutV(oid, schema.ControllerName, "Alpha")

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalDeviceName(t *testing.T) {
	h := []logs.LogEntry{}
	for _, v := range history {
		h = append(h, v)
	}

	sys.logs.Logs = logs.NewLogs(h...)

	oid := schema.OID("0.2.1")
	expected := "Alpha7"

	catalog.PutT(catalog.CatalogController{DeviceID: 405419896}, oid)
	catalog.PutV(oid, schema.ControllerName, "Alpha")

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDefaultCardName(t *testing.T) {
	sys.logs.Logs = logs.NewLogs()

	expected := ""

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupCardName(t *testing.T) {
	sys.logs.Logs = logs.NewLogs()

	oid := schema.OID("0.4.1")
	card := cards.Card{
		CatalogCard: catalog.CatalogCard{
			OID: oid,
		},
	}
	expected := "FredF"

	catalog.PutT(card.CatalogCard, oid)
	catalog.PutV(oid, schema.CardNumber, uint32(8165538))
	catalog.PutV(oid, schema.CardName, "FredF")

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalCardName(t *testing.T) {
	h := []logs.LogEntry{}
	for _, v := range history {
		h = append(h, v)
	}

	sys.logs.Logs = logs.NewLogs(h...)

	oid := schema.OID("0.4.1")
	card := cards.Card{
		CatalogCard: catalog.CatalogCard{
			OID: oid,
		},
	}
	expected := "Barney"

	catalog.PutT(card.CatalogCard, oid)
	catalog.PutV(oid, schema.CardNumber, uint32(8165538))
	catalog.PutV(oid, schema.CardName, "FredF")

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDefaultDoorName(t *testing.T) {
	sys.logs.Logs = logs.NewLogs()

	expected := ""

	name := eventDoor(event)
	if name != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDoorName(t *testing.T) {
	sys.logs.Logs = logs.NewLogs()

	controller := schema.OID("0.2.1")
	oid := schema.OID("0.3.1")
	door := doors.Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.1",
		},
	}

	catalog.PutT(catalog.CatalogController{DeviceID: 405419896}, controller)
	catalog.PutV(controller, schema.ControllerName, "Alpha")
	catalog.PutV(controller, schema.ControllerDeviceID, 405419896)
	catalog.PutV(controller, schema.ControllerDoor3, oid)

	catalog.PutT(door.CatalogDoor, oid)
	catalog.PutV(oid, schema.DoorName, "Gringotts")

	expected := "Gringotts"

	name := eventDoor(event)
	if name != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalDoorName(t *testing.T) {
	h := []logs.LogEntry{}
	for _, v := range history {
		h = append(h, v)
	}

	sys.logs.Logs = logs.NewLogs(h...)

	controller := schema.OID("0.2.1")
	oid := schema.OID("0.3.1")
	door := doors.Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: oid,
		},
	}

	catalog.PutT(catalog.CatalogController{DeviceID: 405419896}, controller)
	catalog.PutV(controller, schema.ControllerName, "Alpha")
	catalog.PutV(controller, schema.ControllerDeviceID, 405419896)
	catalog.PutV(controller, schema.ControllerDoor3, oid)

	catalog.PutT(door.CatalogDoor, oid)
	catalog.PutV(oid, schema.DoorName, "Gringotts")

	expected := "Cupboard"

	name := eventDoor(event)
	if name != expected {
		t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
	}
}
