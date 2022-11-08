package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/log"
)

type login struct {
	id      uuid.UUID
	touched time.Time
}

type session struct {
	id      uuid.UUID
	touched time.Time
}

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
	VerifyAuthHeader(authorization string) error
	Options(uid string) options
}

func debugf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Debugf("%v", args...)
	} else {
		log.Debugf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func infof(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Infof("%v", args...)
	} else {
		log.Infof(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func warnf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Warnf("%v", args...)
	} else {
		log.Warnf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}
