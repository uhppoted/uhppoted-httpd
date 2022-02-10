package users

import (
	"log"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/types"
)

const GZIP_MINIMUM = 16384

func Get(auth auth.OpAuth) interface{} {
	return struct {
		Users interface{} `json:"users"`
	}{
		Users: system.Users(auth),
	}
}

func Post(body map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	updated, err := system.UpdateUsers(body, auth)
	if err != nil {
		return nil, err
	}

	return struct {
		Users interface{} `json:"users"`
	}{
		Users: updated,
	}, nil
}

func warn(err error) {
	switch v := err.(type) {
	case *types.HttpdError:
		log.Printf("%-5s %v", "WARN", v.Detail)

	default:
		log.Printf("%-5s %v", "WARN", v)
	}
}
