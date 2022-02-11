package groups

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) interface{} {
	return struct {
		Groups interface{} `json:"groups"`
	}{
		Groups: system.Groups(uid, role),
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	updated, err := system.UpdateGroups(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Groups interface{} `json:"groups"`
	}{
		Groups: updated,
	}, nil
}
