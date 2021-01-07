package types

import ()

type Door struct {
	ID           string `json:"id"`
	ControllerID uint32 `json:"device-id"`
	Door         uint8  `json:"door"`
	Name         string `json:"name"`
}

func (d *Door) Clone() Door {
	return Door{
		ID:           d.ID,
		ControllerID: d.ControllerID,
		Door:         d.Door,
		Name:         d.Name,
	}
}
