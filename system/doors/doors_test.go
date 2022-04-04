package doors

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestValidateWithMissingDoorName(t *testing.T) {
	dd := Doors{
		doors: map[schema.OID]Door{
			"0.3.5": Door{
				CatalogDoor: catalog.CatalogDoor{
					OID: "0.3.5",
				},
				Name: "",
			},
		},
	}

	if err := dd.Validate(); err == nil {
		t.Errorf("Expected error validating doors list with invalid door, got:%v", err)
	}
}
