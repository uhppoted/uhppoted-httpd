package doors

import (
	"fmt"
)

type info struct {
	DeviceID    string `json:"device-id"`
	DoorID      string `json:"door-id"`
	Door        string `json:"door"`
	FieldName   string `json:"field"`
	Description string `json:"description"`
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
	return i.Description
}
