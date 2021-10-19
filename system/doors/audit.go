package doors

import (
	"fmt"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type info struct {
	OID       catalog.OID `json:"OID"`
	DeviceID  string      `json:"device-id"`
	DoorID    string      `json:"door-id"`
	Door      string      `json:"door"`
	FieldName string      `json:"field"`
	Current   string      `json:"current"`
	Updated   string      `json:"new"`
}

func (i info) ID() string {
	return fmt.Sprintf("%v:%v", i.DeviceID, i.DoorID)
}

func (i info) Name() string {
	return i.Door
}

func (i info) Field() string {
	return i.FieldName
}

func (i info) Details() string {
	switch strings.ToLower(i.FieldName) {
	case "name":
		return fmt.Sprintf("Updated name from %v to %v", i.Current, i.Updated)

	case "delay":
		return fmt.Sprintf("Updated delay from %vs to %vs", i.Current, i.Updated)

	case "mode":
		return fmt.Sprintf("Updated mode from %v to %v", i.Current, i.Updated)

	default:
		return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
	}
}
