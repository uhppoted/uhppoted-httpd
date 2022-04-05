package doors

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestValidateWithInvalid(t *testing.T) {
	dd := Doors{
		doors: map[schema.OID]Door{
			"0.3.5": Door{
				CatalogDoor: catalog.CatalogDoor{
					OID: "0.3.5",
				},
				Name:     "",
				created:  types.TimestampNow(),
				modified: types.TimestampNow(),
			},
		},
	}

	if err := dd.Validate(); err == nil {
		t.Errorf("Expected error validating doors list with invalid door (%v)", err)
	}
}

func TestValidateWithNewDoor(t *testing.T) {
	dd := Doors{
		doors: map[schema.OID]Door{
			"0.3.5": Door{
				CatalogDoor: catalog.CatalogDoor{
					OID: "0.3.5",
				},
				Name:    "",
				created: types.TimestampNow(),
			},
		},
	}

	if err := dd.Validate(); err != nil {
		t.Errorf("Unexpected error validating doors list with invalid door (%v)", err)
	}
}
