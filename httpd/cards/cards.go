package cards

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get() interface{} {
	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: system.Cards(),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	updated, err := system.UpdateCards(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: updated,
	}, nil
}
