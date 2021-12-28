package doors

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(auth auth.OpAuth) interface{} {
	return struct {
		Doors interface{} `json:"doors"`
	}{
		Doors: system.Doors(auth),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	if updated, err := system.UpdateDoors(body, auth); err != nil {
		return nil, err
	} else {
		return struct {
			Doors interface{} `json:"doors"`
		}{
			Doors: updated,
		}, nil
	}
}
