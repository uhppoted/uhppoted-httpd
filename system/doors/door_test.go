package doors

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestDoorAsObjects(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	d := Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.3",
		},
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		created:   created,
	}

	expected := []schema.Object{
		{OID: "0.3.3", Value: ""},
		{OID: "0.3.3.0.0", Value: types.StatusOk},
		{OID: "0.3.3.0.1", Value: created},
		{OID: "0.3.3.0.2", Value: types.Timestamp{}},
		{OID: "0.3.3.1", Value: "Le Door"},
		{OID: "0.3.3.2", Value: types.Uint8(0)},
		{OID: "0.3.3.2.1", Value: types.StatusUnknown},
		{OID: "0.3.3.2.2", Value: uint8(7)},
		{OID: "0.3.3.2.3", Value: ""},
		{OID: "0.3.3.3", Value: core.ControlState(0)},
		{OID: "0.3.3.3.1", Value: types.StatusUnknown},
		{OID: "0.3.3.3.2", Value: core.NormallyOpen},
		{OID: "0.3.3.3.3", Value: ""},
		{OID: "0.3.3.4", Value: true},
		{OID: "0.3.3.5", Value: "******"},
	}

	objects := d.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()

	d := Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.3",
		},
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		created:   created,
		deleted:   deleted,
	}

	expected := []schema.Object{
		{OID: "0.3.3.0.2", Value: deleted},
	}

	objects := d.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	d := Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.3",
		},
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		created:   created,
	}

	expected := []schema.Object{
		{OID: "0.3.3", Value: ""},
		{OID: "0.3.3.0.0", Value: types.StatusOk},
		{OID: "0.3.3.0.1", Value: created},
		{OID: "0.3.3.0.2", Value: types.Timestamp{}},
		{OID: "0.3.3.1", Value: "Le Door"},
		{OID: "0.3.3.3", Value: core.ControlState(0)},
		{OID: "0.3.3.3.1", Value: types.StatusUnknown},
		{OID: "0.3.3.3.2", Value: core.NormallyOpen},
		{OID: "0.3.3.3.3", Value: ""},
		{OID: "0.3.3.4", Value: true},
		{OID: "0.3.3.5", Value: "******"},
	}

	a := auth.Authorizator{
		OpAuth: &stub{
			canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
				if strings.HasPrefix(field, "door.delay") {
					return errors.New("test")
				}

				return nil
			},
		},
	}

	objects := d.AsObjects(&a)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorSet(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.1", Value: "Eine Kleine Dooren"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.3",
		},
		name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.name != "Eine Kleine Dooren" {
		t.Errorf("Door name not updated - expected:%v, got:%v", "Eine Kleine Dooren", d.name)
	}
}

func TestDoorSetWithDeleted(t *testing.T) {
	d := Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: "0.3.3",
		},
		name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,

		deleted: types.TimestampNow(),
	}

	expected := []schema.Object{
		schema.Object{OID: "0.3.3.0.2", Value: d.deleted},
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", db.DBC{})
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.name != "Le Door" {
		t.Errorf("Door name unexpectedly updated - expected:%v, got:%v", "Le Door", d.name)
	}
}
