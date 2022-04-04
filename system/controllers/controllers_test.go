package controllers

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

func TestValidateWithInvalidController(t *testing.T) {
	cc := Controllers{
		controllers: []*Controller{
			&Controller{
				CatalogController: catalog.CatalogController{
					OID:      "0.2.1",
					DeviceID: 0,
				},
				name: "",
			},
		},
	}

	if err := cc.Validate(); err == nil {
		t.Errorf("Expected error validating controllers list with invalid controller, got:%v", err)
	}
}
