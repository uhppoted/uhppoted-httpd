package logs

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestLogEntryAsObjects(t *testing.T) {
	timestamp := time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local)

	l := LogEntry{
		OID:       "0.7.3",
		Timestamp: timestamp,
		UID:       "admin",
		Item:      "thing1",
		ItemID:    "12.34",
		ItemName:  "A Thyngge",
		Field:     "widget",
		Details:   "grokked the widget thing",
		Before:    "wodget",
		After:     "gadget",
	}

	expected := []interface{}{
		catalog.Object{OID: "0.7.3", Value: types.StatusOk},
		catalog.Object{OID: "0.7.3.1", Value: "2021-02-28T12:34:56-08:00"},
		catalog.Object{OID: "0.7.3.2", Value: "admin"},
		catalog.Object{OID: "0.7.3.3", Value: "thing1"},
		catalog.Object{OID: "0.7.3.4", Value: "12.34"},
		catalog.Object{OID: "0.7.3.5", Value: "A Thyngge"},
		catalog.Object{OID: "0.7.3.6", Value: "widget"},
		catalog.Object{OID: "0.7.3.7", Value: "grokked the widget thing"},
	}

	objects := l.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestLogEntryAsObjectsWithAuth(t *testing.T) {
	timestamp := time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local)

	l := LogEntry{
		OID:       "0.7.3",
		Timestamp: timestamp,
		UID:       "admin",
		Item:      "thing1",
		ItemID:    "12.34",
		ItemName:  "A Thyngge",
		Field:     "widget",
		Details:   "grokked the widget thing",
		Before:    "wodget",
		After:     "gadget",
	}

	expected := []interface{}{
		catalog.Object{OID: "0.7.3", Value: types.StatusOk},
		catalog.Object{OID: "0.7.3.1", Value: "2021-02-28T12:34:56-08:00"},
		//		catalog.Object{OID: "0.7.3.2", Value: "admin"},
		catalog.Object{OID: "0.7.3.3", Value: "thing1"},
		catalog.Object{OID: "0.7.3.4", Value: "12.34"},
		catalog.Object{OID: "0.7.3.5", Value: "A Thyngge"},
		catalog.Object{OID: "0.7.3.6", Value: "widget"},
		catalog.Object{OID: "0.7.3.7", Value: "grokked the widget thing"},
	}

	auth := stub{
		canView: func(ruleset string, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "log.UID") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := l.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}
