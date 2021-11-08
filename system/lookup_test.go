package system

import (
	"crypto/sha1"
	"testing"
	"time"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
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
}

func TestLookupDefaultDeviceName(t *testing.T) {
	sys.logs = logs.NewLogs()

	expected := ""

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDeviceNameWithoutRelevantLogs(t *testing.T) {
	sys.logs = logs.NewLogs()

	oid := catalog.OID("0.3.1")
	expected := "Alpha"

	catalog.PutController(405419896, oid)
	catalog.PutV(oid.Append(catalog.ControllerName), "Alpha")

	for k, v := range history {
		if v.Timestamp.Before(time.Time(event.Timestamp)) {
			sys.logs.Logs[k] = v
		}
	}

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalDeviceName(t *testing.T) {
	sys.logs = logs.NewLogs()

	oid := catalog.OID("0.3.1")
	expected := "Alpha7"

	catalog.PutController(405419896, oid)
	catalog.PutV(oid.Append(catalog.ControllerName), "Alpha")

	for k, v := range history {
		sys.logs.Logs[k] = v
	}

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDefaultCardName(t *testing.T) {
	sys.logs = logs.NewLogs()

	expected := ""

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupCardName(t *testing.T) {
	sys.logs = logs.NewLogs()

	oid := catalog.OID("0.3.1")
	expected := "FredF"

	catalog.PutCard(oid)
	catalog.PutV(oid.Append(catalog.CardNumber), uint32(8165538))
	catalog.PutV(oid.Append(catalog.CardName), "FredF")

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalCardName(t *testing.T) {
	sys.logs = logs.NewLogs()

	oid := catalog.OID("0.3.1")
	expected := "Barney"

	catalog.PutCard(oid)
	catalog.PutV(oid.Append(catalog.CardNumber), uint32(8165538))
	catalog.PutV(oid.Append(catalog.CardName), "FredF")

	for k, v := range history {
		sys.logs.Logs[k] = v
	}

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}
