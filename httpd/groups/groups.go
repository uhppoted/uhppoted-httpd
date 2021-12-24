package groups

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get() interface{} {
	return struct {
		Groups interface{} `json:"groups"`
	}{
		Groups: system.Groups(),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	updated, err := system.UpdateGroups(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Groups interface{} `json:"groups"`
	}{
		Groups: updated,
	}, nil
}
