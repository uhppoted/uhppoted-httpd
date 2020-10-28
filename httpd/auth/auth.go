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
	Authorized(w http.ResponseWriter, r *http.Request, path string) (string, bool)
	Authenticate(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Session(r *http.Request) (*session, error)
	Sweep()
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
