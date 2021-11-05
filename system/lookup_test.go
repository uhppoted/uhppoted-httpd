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

var history = logs.NewLogs()

func init() {
	hash := func(s string) [20]byte {
		return sha1.Sum([]byte(s))
	}

	history.Logs[hash("CONTROLLER.1")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 1, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha1",
		After:     "Alpha2",
	}

	history.Logs[hash("CONTROLLER.2")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 17, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha3",
		After:     "Alpha4",
	}

	history.Logs[hash("CONTROLLER.3")] = logs.LogEntry{
		Timestamp: time.Date(2021, time.October, 27, 12, 34, 15, 0, time.Local),
		Item:      "controller",
		ItemID:    "405419896",
		Field:     "name",
		Before:    "Alpha5",
		After:     "Alpha6",
	}
}

func TestLookupDefaultDeviceName(t *testing.T) {
	expected := ""

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDeviceName(t *testing.T) {
	oid := catalog.OID("0.3.1")
	expected := "Alpha"

	catalog.PutController(405419896, oid)
	catalog.PutV(oid.Append(catalog.ControllerName), "Alpha")
	sys.logs = logs.NewLogs()

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupHistoricalDeviceName(t *testing.T) {
	oid := catalog.OID("0.3.1")
	expected := "Alpha5"

	catalog.PutController(405419896, oid)
	catalog.PutV(oid.Append(catalog.ControllerName), "Alpha")
	sys.logs = history

	name := eventController(event)
	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDefaultCardName(t *testing.T) {
	expected := ""

	name := eventCard(event)
	if name != expected {
		t.Errorf("incorrect card name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupCardName(t *testing.T) {
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
