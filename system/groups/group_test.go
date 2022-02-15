package groups

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

func TestGroupDeserialize(t *testing.T) {
	created = types.Timestamp(time.Date(2022, time.April, 1, 0, 0, 0, 0, time.Local))

	encoded := `{ "OID":"0.5.3", "name":"Le Groupe", "doors":["0.3.3","0.3.7"], "created":"2022-04-01 00:00:00" }`
	expected := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
	}

	var g Group

	if err := g.deserialize([]byte(encoded)); err != nil {
		t.Fatalf("Error deserializing group (%v)", err)
	}

	if !reflect.DeepEqual(g, expected) {
		t.Errorf("Group incorrectly deserialized\n   expected:%#v\n   got:     %#v", expected, g)
	}
}

func TestGroupDeserializeWithDefaultCreated(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	encoded := `{ "OID":"0.5.3", "name":"Le Groupe", "doors":["0.3.3","0.3.7"] }`
	expected := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created.Add(1 * time.Minute),
	}

	var g Group

	if err := g.deserialize([]byte(encoded)); err != nil {
		t.Fatalf("Error deserializing group (%v)", err)
	}

	if !reflect.DeepEqual(g, expected) {
		t.Errorf("Group incorrectly deserialized\n   expected:%#v\n   got:     %#v", expected, g)
		t.Errorf("%v %v", expected.created, g.created)
	}
}

func TestGroupAsObjects(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
	}

	expected := []catalog.Object{
		{OID: "0.5.3", Value: ""},
		{OID: "0.5.3.0.0", Value: types.StatusOk},
		{OID: "0.5.3.0.1", Value: created},
		{OID: "0.5.3.0.2", Value: types.Timestamp{}},
		{OID: "0.5.3.1", Value: "Le Groupe"},
	}

	objects := g.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestGroupAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()

	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
		deleted: deleted,
	}

	expected := []catalog.Object{
		{OID: "0.5.3.0.2", Value: deleted},
	}

	objects := g.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestGroupAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
	}

	expected := []catalog.Object{
		{OID: "0.5.3", Value: ""},
		{OID: "0.5.3.0.0", Value: types.StatusOk},
		{OID: "0.5.3.0.1", Value: created},
		{OID: "0.5.3.0.2", Value: types.Timestamp{}},
		//	{OID: "0.5.3.1", Value: "Le Groupe"},
	}

	auth := stub{
		canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "group.name") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := g.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestGroupSet(t *testing.T) {
	expected := []catalog.Object{
		{OID: "0.5.3", Value: ""},
		{OID: "0.5.3.1", Value: "Ze Gruppe"},
		{OID: "0.5.3.0.0", Value: types.StatusOk},
	}

	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
	}

	objects, err := g.set(nil, "0.5.3.1", "Ze Gruppe", nil)
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if g.Name != "Ze Gruppe" {
		t.Errorf("Group name not updated - expected:%v, got:%v", "Ze Gruppe", g.Name)
	}
}

func TestGroupSetWithDeleted(t *testing.T) {
	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},

		deleted: types.TimestampNow(),
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.5.3.0.2", Value: g.deleted},
	}

	objects, err := g.set(nil, "0.5.3.1", "Ze Gruppe", nil)
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if g.Name != "Le Groupe" {
		t.Errorf("Group name unexpectedly updated - expected:%v, got:%v", "Le Group", g.Name)
	}
}
