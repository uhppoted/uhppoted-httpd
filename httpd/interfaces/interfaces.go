package interfaces

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(auth auth.OpAuth) interface{} {
	return struct {
		Interfaces interface{} `json:"interfaces"`
	}{
		Interfaces: system.Interfaces(auth),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	updated, err := system.UpdateInterfaces(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Interfaces interface{} `json:"interfaces"`
	}{
		Interfaces: updated,
	}, nil
}
