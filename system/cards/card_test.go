package cards

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestCardAsObjects(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	name := types.Name("Le Card")
	card := types.Card(8165537)
	from := types.Date(time.Date(2021, time.March, 1, 0, 0, 056, 0, time.Local))
	to := types.Date(time.Date(2023, time.December, 31, 23, 59, 59, 999, time.Local))

	c := Card{
		OID:     "0.4.3",
		Name:    &name,
		Card:    &card,
		From:    &from,
		To:      &to,
		created: created,
	}

	expected := []interface{}{
		catalog.Object{OID: "0.4.3", Value: ""},
		catalog.Object{OID: "0.4.3.0.0", Value: types.StatusOk},
		catalog.Object{OID: "0.4.3.0.1", Value: created.Format("2006-01-02 15:04:05")},
		catalog.Object{OID: "0.4.3.0.2", Value: (*types.DateTime)(nil)},
		catalog.Object{OID: "0.4.3.1", Value: &name},
		catalog.Object{OID: "0.4.3.2", Value: &card},
		catalog.Object{OID: "0.4.3.3", Value: &from},
		catalog.Object{OID: "0.4.3.4", Value: &to},
	}

	objects := c.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardAsObjectsWithDeleted(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.DateTimePtrNow()
	name := types.Name("Le Card")
	card := types.Card(8165537)
	from := types.Date(time.Date(2021, time.March, 1, 0, 0, 056, 0, time.Local))
	to := types.Date(time.Date(2023, time.December, 31, 23, 59, 59, 999, time.Local))

	c := Card{
		OID:     "0.4.3",
		Name:    &name,
		Card:    &card,
		From:    &from,
		To:      &to,
		created: created,
		deleted: deleted,
	}

	expected := []interface{}{
		catalog.Object{
			OID:   "0.4.3.0.2",
			Value: deleted,
		},
	}

	objects := c.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardSet(t *testing.T) {
	name := types.Name("Le Carte")
	expected := []catalog.Object{
		catalog.Object{OID: "0.4.3.1", Value: "Ze Kardt"},
		catalog.Object{OID: "0.4.3.0.0", Value: types.StatusOk},
	}

	c := Card{
		OID:  "0.4.3",
		Name: &name,
	}

	objects, err := c.set(nil, "0.4.3.1", "Ze Kardt", nil)
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if fmt.Sprintf("%v", c.Name) != "Ze Kardt" {
		t.Errorf("Card name not updated - expected:%v, got:%v", "Ze Kardt", c.Name)
	}
}

func TestCardSetWithDeleted(t *testing.T) {
	name := types.Name("Le Carte")

	c := Card{
		OID:  "0.4.3",
		Name: &name,

		deleted: types.DateTimePtrNow(),
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.4.3.0.2", Value: c.deleted},
	}

	objects, err := c.set(nil, "0.4.3.1", "Ze Kardt", nil)
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if fmt.Sprintf("%v", c.Name) != "Le Carte" {
		t.Errorf("Card name unexpectedly updated - expected:%v, got:%v", "Le Carte", c.Name)
	}
}
