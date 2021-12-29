package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(auth auth.OpAuth) interface{} {
	return struct {
		Controllers interface{} `json:"controllers"`
	}{
		Controllers: system.Controllers(auth),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	updated, err := system.UpdateControllers(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Controllers interface{} `json:"controllers"`
	}{
		Controllers: updated,
	}, nil
}
