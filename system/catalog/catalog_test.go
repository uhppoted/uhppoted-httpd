package catalog

import (
	"testing"
)

func TestNewEvent(t *testing.T) {
	tests := []OID{
		OID("0.6.1"),
		OID("0.6.2"),
		OID("0.6.3"),
	}

	for _, expected := range tests {
		oid := NewEvent()

		if oid != expected {
			t.Errorf("Invalid event OID - expected:%v, got:%v", expected, oid)
		}
	}
}
