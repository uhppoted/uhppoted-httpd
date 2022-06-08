package users

import (
	"log"

	"github.com/uhppoted/uhppoted-httpd/system"
)

const GZIP_MINIMUM = 16384

func Get(uid, role string) interface{} {
	return struct {
		Users interface{} `json:"users"`
	}{
		Users: system.Users(uid, role),
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	updated, err := system.UpdateUsers(uid, role, body)
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
	log.Printf("%-5s %v", "WARN", err)
}
