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