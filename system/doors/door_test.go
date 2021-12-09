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
