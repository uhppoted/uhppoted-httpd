package groups

import (
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestGroupDeserialize(t *testing.T) {
	created = types.DateTime(time.Date(2022, time.April, 1, 0, 0, 0, 0, time.Local))

	encoded := `{ "OID":"0.5.3", "name":"Le Groupe", "doors":["0.3.3","0.3.7"], "created":"2022-04-01 00:00:00" }`
	expected := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
		deleted: nil,
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
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	encoded := `{ "OID":"0.5.3", "name":"Le Groupe", "doors":["0.3.3","0.3.7"] }`
	expected := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created.Add(1 * time.Minute),
		deleted: nil,
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
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	g := Group{
		OID:  "0.5.3",
		Name: "Le Groupe",
		Doors: map[catalog.OID]bool{
			"0.3.3": true,
			"0.3.7": true,
		},
		created: created,
	}

	expected := []interface{}{
		catalog.Object{OID: "0.5.3", Value: ""},
		catalog.Object{OID: "0.5.3.0.0", Value: types.StatusOk},
		catalog.Object{OID: "0.5.3.0.1", Value: created.Format("2006-01-02 15:04:05")},
		catalog.Object{OID: "0.5.3.0.2", Value: (*types.DateTime)(nil)},
		catalog.Object{OID: "0.5.3.1", Value: "Le Groupe"},
	}

	objects := g.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestGroupAsObjectsWithDeleted(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.DateTimePtrNow()

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

	expected := []interface{}{
		catalog.Object{
			OID:   "0.5.3.0.2",
			Value: deleted,
		},
	}

	objects := g.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}
