package users

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(auth auth.OpAuth) interface{} {
	return struct {
		Users interface{} `json:"users"`
	}{
		Users: system.Users(auth),
	}
}
