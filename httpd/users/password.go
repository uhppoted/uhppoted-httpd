package users

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Password(body map[string]interface{}, role string, auth auth.IAuth) (interface{}, error) {
	var uid string
	var old string
	var pwd string

	f := func(k string) (string, error) {
		if v, ok := body[k]; !ok {
			return "", fmt.Errorf("Invalid user ID or password")
		} else if u, ok := v.([]string); ok && len(u) > 0 {
			return u[0], nil
		} else if u, ok := v.(string); ok {
			return u, nil
		}

		return "", fmt.Errorf("Invalid user ID or password")
	}

	if v, err := f("uid"); err != nil {
		return nil, err
	} else {
		uid = v
	}

	if v, err := f("old"); err != nil {
		return nil, err
	} else {
		old = v
	}

	if v, err := f("pwd"); err != nil {
		return nil, err
	} else {
		pwd = v
	}

	// ... validate
	if err := auth.Verify(uid, old); err != nil {
		return nil, fmt.Errorf("Invalid user ID or password")
	}

	if err := system.SetPassword(uid, role, pwd); err != nil {
		return nil, err
	}

	return struct{}{}, nil
}
