package auth

import (
	"fmt"
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/log"
)

type options struct {
	OTP struct {
		Allowed bool
		Enabled bool
	}
}

type IAuth interface {
	Preauthenticate() (*http.Cookie, error)
	Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error)
	Authenticated(cookie *http.Cookie) (string, string, *http.Cookie, error)
	Authorised(uid, role, path string) error
	Logout(cookie *http.Cookie)

	Verify(uid, pwd string) error
	VerifyAuthHeader(uid string, header string) error
	Options(uid, role string) options

	AdminRole() string
}

func warnf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Warnf("%v", args...)
	} else {
		log.Warnf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}
