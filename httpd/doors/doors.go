package doors

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) any {
	return struct {
		Doors any `json:"doors"`
	}{
		Doors: system.Doors(uid, role),
	}
}

func Post(uid, role string, body map[string]any) (any, error) {
	if updated, err := system.UpdateDoors(uid, role, body); err != nil {
		return nil, err
	} else {
		return struct {
			Doors any `json:"doors"`
		}{
			Doors: updated,
		}, nil
	}
}
