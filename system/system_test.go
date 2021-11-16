package system

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

func TestSquooshWithoutDuplicates(t *testing.T) {
	objects := []catalog.Object{
		catalog.Object{OID: "0.3.1", Value: "A"},
		catalog.Object{OID: "0.3.2", Value: "b"},
		catalog.Object{OID: "0.3.3", Value: "C"},
		catalog.Object{OID: "0.3.4", Value: "D"},
		catalog.Object{OID: "0.3.5", Value: "E"},
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1", Value: "A"},
		catalog.Object{OID: "0.3.2", Value: "b"},
		catalog.Object{OID: "0.3.3", Value: "C"},
		catalog.Object{OID: "0.3.4", Value: "D"},
		catalog.Object{OID: "0.3.5", Value: "E"},
	}

	list := squoosh(objects)

	if !reflect.DeepEqual(list, expected) {
		t.Errorf("Incorrectly squoooshed list:\n   expected:%v\n   got:     %v", expected, list)
	}
}

func TestSquooshWithDuplicates(t *testing.T) {
	objects := []catalog.Object{
		catalog.Object{OID: "0.3.1", Value: "A"},
		catalog.Object{OID: "0.3.2", Value: "b"},
		catalog.Object{OID: "0.3.3", Value: "C"},
		catalog.Object{OID: "0.3.2", Value: "D"},
		catalog.Object{OID: "0.3.5", Value: "E"},
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1", Value: "A"},
		catalog.Object{OID: "0.3.3", Value: "C"},
		catalog.Object{OID: "0.3.2", Value: "D"},
		catalog.Object{OID: "0.3.5", Value: "E"},
	}

	list := squoosh(objects)

	if !reflect.DeepEqual(list, expected) {
		t.Errorf("Incorrectly squoooshed list:\n   expected:%v\n   got:     %v", expected, list)
	}
}
