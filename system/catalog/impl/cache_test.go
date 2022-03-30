package memdb

import (
	"testing"

	cat "github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

func TestGetDoorDeviceID(t *testing.T) {
	db := NewCatalog()
	door := schema.OID("0.3.5")

	db.PutT(cat.CatalogController{DeviceID: 405419896}, "0.2.7")
	db.PutV("0.2.7", schema.ControllerDoor3, door)

	d := db.GetDoorDeviceID(door)

	if d != 405419896 {
		t.Errorf("Incorrect device ID - expected:%v, got:%v", 405419896, d)
	}
}

func TestGetDoorDeviceDoor(t *testing.T) {
	db := NewCatalog()
	door := schema.OID("0.3.5")

	db.PutT(cat.CatalogController{DeviceID: 405419896}, "0.2.7")
	db.PutV("0.2.7", schema.ControllerDoor3, door)

	d := db.GetDoorDeviceDoor(door)

	if d != 3 {
		t.Errorf("Incorrect device door - expected:%v, got:%v", 3, d)
	}
}
