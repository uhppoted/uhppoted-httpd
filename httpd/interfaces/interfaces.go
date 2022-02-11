package interfaces

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) interface{} {
	return struct {
		Interfaces interface{} `json:"interfaces"`
	}{
		Interfaces: system.Interfaces(uid, role),
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	updated, err := system.UpdateInterfaces(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Interfaces interface{} `json:"interfaces"`
	}{
		Interfaces: updated,
	}, nil
}
