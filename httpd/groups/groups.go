package groups

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) any {
	return struct {
		Groups any `json:"groups"`
	}{
		Groups: system.Groups(uid, role),
	}
}

func Post(uid, role string, body map[string]any) (any, error) {
	updated, err := system.UpdateGroups(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Groups any `json:"groups"`
	}{
		Groups: updated,
	}, nil
}
