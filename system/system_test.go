package system

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestSquooshWithoutDuplicates(t *testing.T) {
	objects := []schema.Object{
		schema.Object{OID: "0.3.1", Value: "A"},
		schema.Object{OID: "0.3.2", Value: "b"},
		schema.Object{OID: "0.3.3", Value: "C"},
		schema.Object{OID: "0.3.4", Value: "D"},
		schema.Object{OID: "0.3.5", Value: "E"},
	}

	expected := []schema.Object{
		schema.Object{OID: "0.3.1", Value: "A"},
		schema.Object{OID: "0.3.2", Value: "b"},
		schema.Object{OID: "0.3.3", Value: "C"},
		schema.Object{OID: "0.3.4", Value: "D"},
		schema.Object{OID: "0.3.5", Value: "E"},
	}

	list := squoosh(objects)

	if !reflect.DeepEqual(list, expected) {
		t.Errorf("Incorrectly squoooshed list:\n   expected:%v\n   got:     %v", expected, list)
	}
}

func TestSquooshWithDuplicates(t *testing.T) {
	objects := []schema.Object{
		schema.Object{OID: "0.3.1", Value: "A"},
		schema.Object{OID: "0.3.2", Value: "b"},
		schema.Object{OID: "0.3.3", Value: "C"},
		schema.Object{OID: "0.3.2", Value: "D"},
		schema.Object{OID: "0.3.5", Value: "E"},
	}

	expected := []schema.Object{
		schema.Object{OID: "0.3.1", Value: "A"},
		schema.Object{OID: "0.3.3", Value: "C"},
		schema.Object{OID: "0.3.2", Value: "D"},
		schema.Object{OID: "0.3.5", Value: "E"},
	}

	list := squoosh(objects)

	if !reflect.DeepEqual(list, expected) {
		t.Errorf("Incorrectly squoooshed list:\n   expected:%v\n   got:     %v", expected, list)
	}
}
