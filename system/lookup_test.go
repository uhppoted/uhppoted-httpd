package system

import (
	"testing"
	"time"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
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
