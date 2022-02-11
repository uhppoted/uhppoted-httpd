package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) interface{} {
	return struct {
		Controllers interface{} `json:"controllers"`
	}{
		Controllers: system.Controllers(uid, role),
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	auth := auth.NewAuthorizator(uid, role)

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
