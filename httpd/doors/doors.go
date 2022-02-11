package doors

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) interface{} {
	return struct {
		Doors interface{} `json:"doors"`
	}{
		Doors: system.Doors(uid, role),
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	if updated, err := system.UpdateDoors(uid, role, body); err != nil {
		return nil, err
	} else {
		return struct {
			Doors interface{} `json:"doors"`
		}{
			Doors: updated,
		}, nil
	}
}
