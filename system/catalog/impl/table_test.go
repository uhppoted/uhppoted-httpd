package memdb

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestTableNewT(t *testing.T) {
	tt := table{
		base: schema.DoorsOID,
		m:    map[schema.OID]record{},
		last: 123,
	}

	expected := schema.OID("0.3.5")
	oid := tt.New(struct{}{})

	if oid != "0.3.124" {
		t.Errorf("Incorrect new OID - expected:%v, got:%v", expected, oid)
	}
}