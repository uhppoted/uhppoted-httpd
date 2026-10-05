package cards

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	lib "codeberg.org/uhppoted/uhppoted-core/types"

	"codeberg.org/uhppoted/uhppoted-httpd/auth"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog/impl"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog/schema"
	"codeberg.org/uhppoted/uhppoted-httpd/system/db"
	"codeberg.org/uhppoted/uhppoted-httpd/types"
)

func TestCardAsObjects(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	from := lib.MustParseDate("2021-03-01")
	to := lib.MustParseDate("2023-12-31")

	c := Card{
		OID:     "0.4.3",
		CardID:  8165537,
		name:    "Le Card",
		pin:     7531,
		from:    from,
		to:      to,
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: types.Timestamp{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		{OID: "0.4.3.2", Value: uint32(8165537)},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
		{OID: "0.4.3.6", Value: uint32(7531)},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardWithMissingFromDateAsObjects(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	from := lib.Date{}
	to := lib.MustParseDate("2023-12-31")

	c := Card{
		OID:     "0.4.3",
		CardID:  8165537,
		name:    "Le Card",
		pin:     7531,
		to:      to,
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusIncomplete},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: types.Timestamp{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		{OID: "0.4.3.2", Value: uint32(8165537)},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
		{OID: "0.4.3.6", Value: uint32(7531)},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardWithMissingToDateAsObjects(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	from := lib.MustParseDate("2021-03-01")
	to := lib.Date{}

	c := Card{
		OID:     "0.4.3",
		CardID:  8165537,
		name:    "Le Card",
		pin:     7531,
		from:    from,
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusIncomplete},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: types.Timestamp{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		{OID: "0.4.3.2", Value: uint32(8165537)},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
		{OID: "0.4.3.6", Value: uint32(7531)},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()
	from := lib.MustParseDate("2021-03-01")
	to := lib.MustParseDate("2023-12-31")

	c := Card{
		OID:     "0.4.3",
		CardID:  8165537,
		name:    "Le Card",
		from:    from,
		to:      to,
		created: created,
		deleted: deleted,
	}

	expected := []schema.Object{
		{OID: "0.4.3.0.2", Value: deleted},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	from := lib.MustParseDate("2021-03-01")
	to := lib.MustParseDate("2023-12-31")

	c := Card{
		OID:     "0.4.3",
		CardID:  8165537,
		name:    "Le Card",
		pin:     7531,
		from:    from,
		to:      to,
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
		{OID: "0.4.3.0.1", Value: created},
		{OID: "0.4.3.0.2", Value: types.Timestamp{}},
		{OID: "0.4.3.1", Value: "Le Card"},
		{OID: "0.4.3.3", Value: from},
		{OID: "0.4.3.4", Value: to},
		{OID: "0.4.3.6", Value: uint32(7531)},
	}

	a := auth.Authorizator{
		OpAuth: &stub{
			canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value any) error {
				if strings.HasPrefix(field, "card.number") {
					return errors.New("test")
				}

				return nil
			},
		},
	}

	objects := c.AsObjects(&a)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestCardSet(t *testing.T) {
	expected := []schema.Object{
		{OID: "0.4.3", Value: ""},
		{OID: "0.4.3.1", Value: "Ze Kardt"},
		{OID: "0.4.3.0.0", Value: types.StatusOk},
	}

	c := Card{
		OID:  "0.4.3",
		name: "Le Carte",
		from: lib.MustParseDate("2024-01-01"),
		to:   lib.MustParseDate("2024-12-31"),
	}

	objects, err := c.set(nil, "0.4.3.1", "Ze Kardt", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if fmt.Sprintf("%v", c.name) != "Ze Kardt" {
		t.Errorf("Card name not updated - expected:%v, got:%v", "Ze Kardt", c.name)
	}
}

func TestCardSetWithDeleted(t *testing.T) {
	c := Card{
		OID:  "0.4.3",
		name: "Le Carte",

		deleted: types.TimestampNow(),
	}

	expected := []schema.Object{
		schema.Object{OID: "0.4.3.0.2", Value: c.deleted},
	}

	objects, err := c.set(nil, "0.4.3.1", "Ze Kardt", db.DBC{})
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if fmt.Sprintf("%v", c.name) != "Le Carte" {
		t.Errorf("Card name unexpectedly updated - expected:%v, got:%v", "Le Carte", c.name)
	}
}
