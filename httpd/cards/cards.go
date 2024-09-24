package cards

import (
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string) interface{} {
	cards := system.Cards(uid, role)

	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: cards,
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	updated, err := system.UpdateCards(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: updated,
	}, nil
}
