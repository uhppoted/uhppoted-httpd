package controllers

import (
	"testing"

	"codeberg.org/uhppoted/uhppoted-httpd/types"
)

func TestValidateWithInvalidController(t *testing.T) {
	cc := Controllers{
		controllers: []*Controller{
			&Controller{
				OID:      "0.2.1",
				DeviceID: 0,
				name:     "",
				created:  types.TimestampNow(),
				modified: types.TimestampNow(),
			},
		},
	}

	if err := cc.Validate(); err == nil {
		t.Errorf("Expected error validating controllers list with invalid controller (%v)", err)
	}
}

func TestValidateWithNewController(t *testing.T) {
	cc := Controllers{
		controllers: []*Controller{
			&Controller{
				OID:      "0.2.1",
				DeviceID: 0,
				name:     "",
				created:  types.TimestampNow(),
			},
		},
	}

	if err := cc.Validate(); err != nil {
		t.Errorf("Unexpected error validating controllers list with new controller (%v)", err)
	}
}
