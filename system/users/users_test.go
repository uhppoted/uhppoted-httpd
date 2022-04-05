package users

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestValidateWithInvalidUser(t *testing.T) {
	uu := Users{
		users: map[schema.OID]*User{
			"0.8.9": &User{
				CatalogUser: catalog.CatalogUser{
					OID: "0.8.9",
				},
				name:     "",
				uid:      "",
				created:  types.TimestampNow(),
				modified: types.TimestampNow(),
			},
		},
	}

	if err := uu.Validate(); err == nil {
		t.Errorf("Expected error validating users list with invalid user (%v)", err)
	}
}

func TestValidateWithNewUser(t *testing.T) {
	uu := Users{
		users: map[schema.OID]*User{
			"0.8.9": &User{
				CatalogUser: catalog.CatalogUser{
					OID: "0.8.9",
				},
				name:    "",
				uid:     "",
				created: types.TimestampNow(),
			},
		},
	}

	if err := uu.Validate(); err != nil {
		t.Errorf("Unexpected error validating users list with new user (%v)", err)
	}
}
