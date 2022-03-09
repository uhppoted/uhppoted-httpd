package catalog

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestJoin(t *testing.T) {
	p := []schema.Object{
		schema.Object{OID: "0.1.1", Value: "A"},
		schema.Object{OID: "0.1.2", Value: "B"},
		schema.Object{OID: "0.1.3", Value: "C"},
	}

	q := []schema.Object{
		schema.Object{OID: "0.2.1", Value: "X"},
		schema.Object{OID: "0.3.2", Value: "Y"},
		schema.Object{OID: "0.3.3", Value: "Z"},
	}

	expected := []schema.Object{
		schema.Object{OID: "0.1.1", Value: "A"},
		schema.Object{OID: "0.1.2", Value: "B"},
		schema.Object{OID: "0.1.3", Value: "C"},
		schema.Object{OID: "0.2.1", Value: "X"},
		schema.Object{OID: "0.3.2", Value: "Y"},
		schema.Object{OID: "0.3.3", Value: "Z"},
	}

	Join(&p, q...)

	if !reflect.DeepEqual(p, expected) {
		t.Errorf("Object lists not joined correctly\n   expected:%v\n   got:     %v", expected, p)
	}
}
