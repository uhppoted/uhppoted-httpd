package system

import (
	"testing"
	"time"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func TestLookupDefaultDeviceName(t *testing.T) {
	timestamp, _ := time.Parse("2006-01-02 15:03:04", "2021-10-26 13:14:15")

	e := uhppoted.Event{
		DeviceID:   405419896,
		Index:      17,
		Type:       1,
		Granted:    true,
		Door:       3,
		Direction:  1,
		CardNumber: 8165538,
		Timestamp:  types.DateTime(timestamp),
		Reason:     1,
	}

	expected := ""
	name := lookup(e)

	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}

func TestLookupDeviceName(t *testing.T) {
	timestamp, _ := time.Parse("2006-01-02 15:03:04", "2021-10-26 13:14:15")
	oid := catalog.OID("0.3.1")

	catalog.PutController(405419896, oid)
	catalog.PutV(oid.Append(catalog.ControllerName), "Alpha", false)

	e := uhppoted.Event{
		DeviceID:   405419896,
		Index:      17,
		Type:       1,
		Granted:    true,
		Door:       3,
		Direction:  1,
		CardNumber: 8165538,
		Timestamp:  types.DateTime(timestamp),
		Reason:     1,
	}

	expected := "Alpha"
	name := lookup(e)

	if name != expected {
		t.Errorf("incorrect device name - expected:%v, got:%v", expected, name)
	}
}
