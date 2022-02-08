package cards

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestCardAsObjects(t *testing.T) {
	created = core.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	card := types.Card(8165537)
	from := core.Date(time.Date(2021, time.March, 1, 0, 0, 056, 0, time.Local))
	to := core.Date(time.Date(2023, time.December, 31, 23, 59, 59, 999, time.Local))

	c := Card{
		OID:     "0.4.3",
		Name:    "Le Card",
		Card:    &card,
		From:    from,
		To:      to,
		created: created,
	}

	expected := []catalog.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: core.DateTime{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		{OID: "0.4.3.2", Value: &card},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardAsObjectsWithDeleted(t *testing.T) {
	created = core.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := core.DateTimeNow()
	card := types.Card(8165537)
	from := core.Date(time.Date(2021, time.March, 1, 0, 0, 056, 0, time.Local))
	to := core.Date(time.Date(2023, time.December, 31, 23, 59, 59, 999, time.Local))

	c := Card{
		OID:     "0.4.3",
		Name:    "Le Card",
		Card:    &card,
		From:    from,
		To:      to,
		created: created,
		deleted: deleted,
	}

	expected := []catalog.Object{
		{OID: "0.4.3.0.2", Value: deleted},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardAsObjectsWithAuth(t *testing.T) {
	created = core.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	card := types.Card(8165537)
	from := core.Date(time.Date(2021, time.March, 1, 0, 0, 056, 0, time.Local))
	to := core.Date(time.Date(2023, time.December, 31, 23, 59, 59, 999, time.Local))

	c := Card{
		OID:     "0.4.3",
		Name:    "Le Card",
		Card:    &card,
		From:    from,
		To:      to,
		created: created,
	}

	expected := []catalog.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: core.DateTime{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		// {OID: "0.4.3.2", Value: &card},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
	}

	auth := stub{
		canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "card.number") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := c.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardSet(t *testing.T) {
	expected := []catalog.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.1", Value: "Ze Kardt"},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
	}

	c := Card{
		OID:  "0.4.3",
		Name: "Le Carte",
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
	c := Card{
		OID:  "0.4.3",
		Name: "Le Carte",

		deleted: core.DateTimeNow(),
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
