package groups

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestValidateWithMissingGroupName(t *testing.T) {
	gg := Groups{
		groups: map[schema.OID]Group{
			"0.5.7": Group{
				CatalogGroup: catalog.CatalogGroup{
					OID: "0.5.7",
				},
				Name: "",
			},
		},
	}

	if err := gg.Validate(); err == nil {
		t.Errorf("Expected error validating groups list with invalid group, got:%v", err)
	}
}
