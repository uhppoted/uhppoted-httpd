package types

import ()

type Door struct {
	ID       string `json:"id"`
	DeviceID uint32 `json:"device-id"`
	Door     uint8  `json:"door"`
	Name     string `json:"name"`
}

func (d *Door) Clone() Door {
	return Door{
		ID:       d.ID,
		DeviceID: d.DeviceID,
		Door:     d.Door,
		Name:     d.Name,
	}
}
