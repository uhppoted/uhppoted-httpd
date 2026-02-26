package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) any {
	return struct {
		Controllers any `json:"controllers"`
	}{
		Controllers: system.Controllers(uid, role),
	}
}

func Post(uid, role string, body map[string]any) (any, error) {
	auth := auth.NewAuthorizator(uid, role)

	updated, err := system.UpdateControllers(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Controllers any `json:"controllers"`
	}{
		Controllers: updated,
	}, nil
}
