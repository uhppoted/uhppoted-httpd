package types

import ()

type Door struct {
	ID           string `json:"id"`
	ControllerID uint32 `json:"device-id"`
	Door         uint8  `json:"door"`
	Name         string `json:"name"`
}
