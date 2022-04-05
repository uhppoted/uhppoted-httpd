package groups

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestValidateWithInvalidGroup(t *testing.T) {
	gg := Groups{
		groups: map[schema.OID]Group{
			"0.5.7": Group{
				CatalogGroup: catalog.CatalogGroup{
					OID: "0.5.7",
				},
				Name:     "",
				created:  types.TimestampNow(),
				modified: types.TimestampNow(),
			},
		},
	}

	if err := gg.Validate(); err == nil {
		t.Errorf("Expected error validating groups list with invalid group (%v)", err)
	}
}

func TestValidateWithNewGroup(t *testing.T) {
	gg := Groups{
		groups: map[schema.OID]Group{
			"0.5.7": Group{
				CatalogGroup: catalog.CatalogGroup{
					OID: "0.5.7",
				},
				Name:    "",
				created: types.TimestampNow(),
			},
		},
	}

	if err := gg.Validate(); err != nil {
		t.Errorf("Unexpected error validating groups list with new group (%v)", err)
	}
}
