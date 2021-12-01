package catalog

import (
	"testing"
)

func TestGetDoorDeviceID(t *testing.T) {
	door := OID("0.3.5")

	PutController(405419896, "0.2.7")
	PutV("0.2.7", ControllerDoor3, door)

	d := GetDoorDeviceID(door)

	if d != 405419896 {
		t.Errorf("Incorrect device ID - expected:%v, got:%v", 405419896, d)
	}
}

func TestGetDoorDeviceDoor(t *testing.T) {
	door := OID("0.3.5")

	PutController(405419896, "0.2.7")
	PutV("0.2.7", ControllerDoor3, door)

	d := GetDoorDeviceDoor(door)

	if d != 3 {
		t.Errorf("Incorrect device door - expected:%v, got:%v", 3, d)
	}
}
