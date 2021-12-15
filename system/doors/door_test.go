package doors

import (
	"reflect"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestDoorAsObjects(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	d := Door{
		OID:     "0.3.3",
		Name:    "Le Door",
		delay:   7,
		mode:    core.NormallyOpen,
		created: created,
	}

	expected := []interface{}{
		catalog.Object{OID: "0.3.3", Value: ""},
		catalog.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
		catalog.Object{OID: "0.3.3.0.1", Value: created.Format("2006-01-02 15:04:05")},
		catalog.Object{OID: "0.3.3.0.2", Value: (*types.DateTime)(nil)},
		catalog.Object{OID: "0.3.3.1", Value: "Le Door"},
		catalog.Object{OID: "0.3.3.2", Value: types.Uint8(0)},
		catalog.Object{OID: "0.3.3.2.1", Value: types.StatusUnknown},
		catalog.Object{OID: "0.3.3.2.2", Value: uint8(7)},
		catalog.Object{OID: "0.3.3.2.3", Value: ""},
		catalog.Object{OID: "0.3.3.3", Value: core.ControlState(0)},
		catalog.Object{OID: "0.3.3.3.1", Value: types.StatusUnknown},
		catalog.Object{OID: "0.3.3.3.2", Value: core.NormallyOpen},
		catalog.Object{OID: "0.3.3.3.3", Value: ""},
	}

	objects := d.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorAsObjectsWithDeleted(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.DateTimePtrNow()

	d := Door{
		OID:     "0.3.3",
		Name:    "Le Door",
		delay:   7,
		mode:    core.NormallyOpen,
		created: created,
		deleted: deleted,
	}

	expected := []interface{}{
		catalog.Object{
			OID:   "0.3.3.0.2",
			Value: deleted,
		},
	}

	objects := d.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorSet(t *testing.T) {
	expected := []catalog.Object{
		catalog.Object{OID: "0.3.3.1", Value: "Eine Kleine Dooren"},
		catalog.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		Name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", nil)
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.Name != "Eine Kleine Dooren" {
		t.Errorf("Door name not updated - expected:%v, got:%v", "Eine Kleine Dooren", d.Name)
	}
}

func TestDoorSetWithDeleted(t *testing.T) {
	d := Door{
		OID:   "0.3.3",
		Name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,

		deleted: types.DateTimePtrNow(),
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.3.0.2", Value: d.deleted},
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", nil)
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.Name != "Le Door" {
		t.Errorf("Door name unexpectedly updated - expected:%v, got:%v", "Le Door", d.Name)
	}
}