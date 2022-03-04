package events

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestEventAsObjects(t *testing.T) {
	timestamp := core.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	e := Event{
		OID:        "0.6.3",
		DeviceID:   405419896,
		Index:      79,
		Timestamp:  timestamp,
		Type:       6,
		Door:       3,
		Direction:  1,
		Card:       8165537,
		Granted:    true,
		Reason:     0x55,
		DeviceName: "Le Controlleur",
		DoorName:   "Ze Door",
		CardName:   "Eine Kardt",
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.6.3", Value: types.StatusOk},
		catalog.Object{OID: "0.6.3.2", Value: uint32(405419896)},
		catalog.Object{OID: "0.6.3.1", Value: timestamp},
		catalog.Object{OID: "0.6.3.4", Value: eventType(6)},
		catalog.Object{OID: "0.6.3.5", Value: uint8(3)},
		catalog.Object{OID: "0.6.3.6", Value: direction(1)},
		catalog.Object{OID: "0.6.3.7", Value: uint32(8165537)},
		catalog.Object{OID: "0.6.3.9", Value: reason(0x55)},
		catalog.Object{OID: "0.6.3.8", Value: true},
		catalog.Object{OID: "0.6.3.10", Value: "Le Controlleur"},
		catalog.Object{OID: "0.6.3.11", Value: "Ze Door"},
		catalog.Object{OID: "0.6.3.12", Value: "Eine Kardt"},
	}

	objects := e.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestEventAsObjectsWithAuth(t *testing.T) {
	timestamp := core.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	e := Event{
		OID:        "0.6.3",
		DeviceID:   405419896,
		Index:      79,
		Timestamp:  timestamp,
		Type:       6,
		Door:       3,
		Direction:  1,
		Card:       8165537,
		Granted:    true,
		Reason:     0x55,
		DeviceName: "Le Controlleur",
		DoorName:   "Ze Door",
		CardName:   "Eine Kardt",
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.6.3", Value: types.StatusOk},
		//		catalog.Object{OID: "0.6.3.2", Value: uint32(405419896)},
		catalog.Object{OID: "0.6.3.1", Value: timestamp},
		catalog.Object{OID: "0.6.3.4", Value: eventType(6)},
		catalog.Object{OID: "0.6.3.5", Value: uint8(3)},
		catalog.Object{OID: "0.6.3.6", Value: direction(1)},
		catalog.Object{OID: "0.6.3.7", Value: uint32(8165537)},
		catalog.Object{OID: "0.6.3.9", Value: reason(0x55)},
		catalog.Object{OID: "0.6.3.8", Value: true},
		catalog.Object{OID: "0.6.3.10", Value: "Le Controlleur"},
		catalog.Object{OID: "0.6.3.11", Value: "Ze Door"},
		catalog.Object{OID: "0.6.3.12", Value: "Eine Kardt"},
	}

	auth := stub{
		canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "event.device.ID") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := e.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}
