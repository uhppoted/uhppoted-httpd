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
	Authenticated(r *http.Request) (string, string, bool)
	Authorised(uid, role, path string) error
	Logout(w http.ResponseWriter, r *http.Request)
	Session(r *http.Request) (*session, error)

	Verify(uid, pwd string) error
	SetPassword(uid, pwd, role string) error

	Sweep()
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
