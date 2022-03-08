package memdb

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestNewEvent(t *testing.T) {
	cc := Catalog()

	tests := []schema.OID{
		schema.OID("0.6.1"),
		schema.OID("0.6.2"),
		schema.OID("0.6.3"),
	}

	for _, expected := range tests {
		oid := cc.NewEvent()

		if oid != expected {
			t.Errorf("Invalid event OID - expected:%v, got:%v", expected, oid)
		}
	}
}
