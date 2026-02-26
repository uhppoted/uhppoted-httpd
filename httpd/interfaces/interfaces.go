package interfaces

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) any {
	return struct {
		Interfaces any `json:"interfaces"`
	}{
		Interfaces: system.Interfaces(uid, role),
	}
}

func Post(uid, role string, body map[string]any) (any, error) {
	updated, err := system.UpdateInterfaces(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Interfaces any `json:"interfaces"`
	}{
		Interfaces: updated,
	}, nil
}
