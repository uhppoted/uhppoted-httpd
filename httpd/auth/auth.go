package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	LoginCookie   = "uhppoted-httpd-login"
	SessionCookie = "uhppoted-httpd-session"
)

type login struct {
	id      uuid.UUID
	touched time.Time
}

type session struct {
	id      uuid.UUID
	touched time.Time

	User string
}

type IAuth interface {
	Preauthenticate() (*http.Cookie, error)
	Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error)
	Authenticated(cookie *http.Cookie) (string, string, error)
	Authorised(uid, role, path string) error
	Logout(cookie *http.Cookie)

	Verify(uid, pwd string) error
	SetPassword(uid, pwd, role string) error

	Sweep()
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
