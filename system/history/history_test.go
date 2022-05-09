package history

import (
	// "crypto/sha1"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	// "github.com/uhppoted/uhppote-core/types"
	// "github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	// "github.com/uhppoted/uhppoted-httpd/system/doors"
	// "github.com/uhppoted/uhppoted-httpd/system/logs"
	// "github.com/uhppoted/uhppoted-lib/uhppoted"
)

// var EVENT = uhppoted.Event{
//     DeviceID:   405419896,
//     Index:      17,
//     Type:       1,
//     Granted:    true,
//     Door:       3,
//     Direction:  1,
//     CardNumber: 8165538,
//     Timestamp:  types.DateTime(time.Date(2021, time.October, 26, 13, 14, 15, 0, time.Local)),
//     Reason:     1,
// }

var logs = []Entry{
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
}

func init() {
	// hash := func(s string) [20]byte {
	//     return sha1.Sum([]byte(s))
	// }

	// LOGS[hash("CONTROLLER.1")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
	//     Item:      "controller",
	//     ItemID:    "405419896",
	//     Field:     "name",
	//     Before:    "Alpha1",
	//     After:     "Alpha2",
	// }

	// LOGS[hash("CONTROLLER.2")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
	//     Item:      "controller",
	//     ItemID:    "405419896",
	//     Field:     "name",
	//     Before:    "Alpha3",
	//     After:     "Alpha4",
	// }

	// LOGS[hash("CONTROLLER.3")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
	//     Item:      "controller",
	//     ItemID:    "405419896",
	//     Field:     "name",
	//     Before:    "Alpha5",
	//     After:     "Alpha6",
	// }

	// LOGS[hash("CONTROLLER.4")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
	//     Item:      "controller",
	//     ItemID:    "405419896",
	//     Field:     "name",
	//     Before:    "Alpha7",
	//     After:     "Alpha8",
	// }

	// LOGS[hash("CARD.1")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
	//     Item:      "card",
	//     ItemID:    "8165538",
	//     Field:     "name",
	//     Before:    "Card1",
	//     After:     "Card2",
	// }

	// LOGS[hash("CARD.2")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
	//     Item:      "card",
	//     ItemID:    "8165538",
	//     Field:     "name",
	//     Before:    "Card3",
	//     After:     "Card4",
	// }

	// LOGS[hash("CARD.3")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
	//     Item:      "card",
	//     ItemID:    "8165538",
	//     Field:     "name",
	//     Before:    "Card5",
	//     After:     "Card6",
	// }

	// LOGS[hash("CARD.4")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
	//     Item:      "card",
	//     ItemID:    "8165538",
	//     Field:     "name",
	//     Before:    "Barney",
	//     After:     "Card8",
	// }

	// LOGS[hash("DOOR.1")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
	//     Item:      "door",
	//     ItemID:    "405419896:3",
	//     Field:     "name",
	//     Before:    "Door1",
	//     After:     "Door2",
	// }

	// LOGS[hash("DOOR.2")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
	//     Item:      "door",
	//     ItemID:    "405419896:3",
	//     Field:     "name",
	//     Before:    "Door3",
	//     After:     "Door4",
	// }

	// LOGS[hash("DOOR.3")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 25, 12, 34, 15, 0, time.Local),
	//     Item:      "door",
	//     ItemID:    "405419896:3",
	//     Field:     "name",
	//     Before:    "Door5",
	//     After:     "Door6",
	// }

	// LOGS[hash("DOOR.4")] = logs.LogEntry{
	//     Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
	//     Item:      "door",
	//     ItemID:    "405419896:3",
	//     Field:     "name",
	//     Before:    "Cupboard",
	//     After:     "Door8",
	// }
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

	history := NewHistory(logs...)
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

	history := NewHistory(logs...)
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

	history := NewHistory(logs...)
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

func TestLookupFashionablyNewController(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := "Alpha4"

	history := NewHistory(logs...)
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

	history := NewHistory(logs...)
	timestamp := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.Local)
	controller := history.LookupController(timestamp, 405419896)

	if controller != expected {
		t.Errorf("incorrect controller name - expected:%v, got:%v", expected, controller)
	}
}

// func TestLookupDefaultCardName(t *testing.T) {
//     sys.logs.Logs = logs.NewLogs()

//     expected := ""

//     timestamp := time.Time(EVENT.Timestamp)
//     card := EVENT.CardNumber
//     name := sys.history.lookupCard(timestamp, card)

//     if name != expected {
//         t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
//     }
// }

// func TestLookupCardName(t *testing.T) {
//     sys.logs.Logs = logs.NewLogs()

//     oid := schema.OID("0.4.1")
//     card := cards.Card{
//         CatalogCard: catalog.CatalogCard{
//             OID: oid,
//         },
//     }
//     expected := "FredF"

//     catalog.PutT(card.CatalogCard)
//     catalog.PutV(oid, schema.CardNumber, uint32(8165538))
//     catalog.PutV(oid, schema.CardName, "FredF")

//     timestamp := time.Time(EVENT.Timestamp)
//     cardNumber := EVENT.CardNumber
//     name := sys.history.lookupCard(timestamp, cardNumber)

//     if name != expected {
//         t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
//     }
// }

// func TestLookupHistoricalCardName(t *testing.T) {
//     h := []logs.LogEntry{}
//     for _, v := range LOGS {
//         h = append(h, v)
//     }

//     sys.logs.Logs = logs.NewLogs(h...)

//     oid := schema.OID("0.4.1")
//     card := cards.Card{
//         CatalogCard: catalog.CatalogCard{
//             OID: oid,
//         },
//     }
//     expected := "Barney"

//     catalog.PutT(card.CatalogCard)
//     catalog.PutV(oid, schema.CardNumber, uint32(8165538))
//     catalog.PutV(oid, schema.CardName, "FredF")

//     timestamp := time.Time(EVENT.Timestamp)
//     cardNumber := EVENT.CardNumber
//     name := sys.history.lookupCard(timestamp, cardNumber)

//     if name != expected {
//         t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
//     }
// }

// func TestLookupDefaultDoorName(t *testing.T) {
//     sys.logs.Logs = logs.NewLogs()

//     expected := ""

//     timestamp := time.Time(EVENT.Timestamp)
//     deviceID := EVENT.DeviceID
//     doorID := EVENT.Door
//     door := sys.history.lookupDoor(timestamp, deviceID, doorID)

//     if door != expected {
//         t.Errorf("incorrect door name - expected:%v, got:%v", expected, door)
//     }
// }

// func TestLookupDoorName(t *testing.T) {
//     catalog.Init(memdb.NewCatalog())

//     sys = system{
//         logs: struct {
//             logs.Logs
//             file string
//             tag  string
//         }{
//             Logs: logs.NewLogs(),
//         },
//     }

//     controller := schema.OID("0.2.1")
//     oid := schema.OID("0.3.1")
//     door := doors.Door{
//         CatalogDoor: catalog.CatalogDoor{
//             OID: "0.3.1",
//         },
//     }

//     catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
//     catalog.PutV(controller, schema.ControllerName, "Alpha")
//     catalog.PutV(controller, schema.ControllerDeviceID, 405419896)
//     catalog.PutV(controller, schema.ControllerDoor3, oid)

//     catalog.PutT(door.CatalogDoor)
//     catalog.PutV(oid, schema.DoorName, "Gringotts")

//     expected := "Gringotts"

//     timestamp := time.Time(EVENT.Timestamp)
//     deviceID := EVENT.DeviceID
//     doorID := EVENT.Door
//     name := sys.history.lookupDoor(timestamp, deviceID, doorID)

//     if name != expected {
//         t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
//     }
// }

// func TestLookupHistoricalDoorName(t *testing.T) {
//     h := []logs.LogEntry{}
//     for _, v := range LOGS {
//         h = append(h, v)
//     }

//     sys.logs.Logs = logs.NewLogs(h...)

//     controller := schema.OID("0.2.1")
//     oid := schema.OID("0.3.1")
//     door := doors.Door{
//         CatalogDoor: catalog.CatalogDoor{
//             OID: oid,
//         },
//     }

//     catalog.PutT(catalog.CatalogController{OID: "0.2.1", DeviceID: 405419896})
//     catalog.PutV(controller, schema.ControllerName, "Alpha")
//     catalog.PutV(controller, schema.ControllerDeviceID, 405419896)
//     catalog.PutV(controller, schema.ControllerDoor3, oid)

//     catalog.PutT(door.CatalogDoor)
//     catalog.PutV(oid, schema.DoorName, "Gringotts")

//     expected := "Cupboard"

//     timestamp := time.Time(EVENT.Timestamp)
//     deviceID := EVENT.DeviceID
//     doorID := EVENT.Door
//     name := sys.history.lookupDoor(timestamp, deviceID, doorID)

//     if name != expected {
//         t.Errorf("incorrect door name - expected:%v, got:%v", expected, name)
//     }
// }
