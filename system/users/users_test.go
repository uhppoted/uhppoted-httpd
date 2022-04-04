package users

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestValidateWithInvalidUser(t *testing.T) {
	uu := Users{
		users: map[schema.OID]*User{
			"0.8.9": &User{
				CatalogUser: catalog.CatalogUser{
					OID: "0.8.9",
				},
				name: "",
				uid:  "",
			},
		},
	}

	if err := uu.Validate(); err == nil {
		t.Errorf("Expected error validating users list with invalid user, got:%v", err)
	}
}
