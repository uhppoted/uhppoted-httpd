package catalog

import (
	"testing"
)

func TestGetDoorDeviceID(t *testing.T) {
	door := OID("0.3.5")
	deviceID := uint32(405419896)

	PutV(door, DoorControllerID, &deviceID)

	d := GetDoorDeviceID(door)

	if d != 405419896 {
		t.Errorf("Incorrect device ID - expected:%v, got:%v", 405419896, d)
	}
}

func TestGetDoorDeviceID2(t *testing.T) {
	door := OID("0.3.5")
	deviceID := uint32(405419896)

	PutV(door, DoorControllerID, deviceID)

	d := GetDoorDeviceID(door)

	if d != 405419896 {
		t.Errorf("Incorrect device ID - expected:%v, got:%v", 405419896, d)
	}
}

func TestGetDoorDeviceDoor(t *testing.T) {
	door := OID("0.3.5")
	deviceID := uint32(405419896)

	PutV(door, DoorControllerID, &deviceID)
	PutV(door, DoorControllerDoor, uint8(7))

	d := GetDoorDeviceDoor(door)

	if d != 7 {
		t.Errorf("Incorrect device door - expected:%v, got:%v", 7, d)
	}
}
