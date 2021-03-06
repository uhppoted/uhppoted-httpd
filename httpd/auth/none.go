package auth

import (
	"net/http"
	"time"
)

type None struct {
	cookieMaxAge int
	stale        time.Duration
}

func NewNoneAuthenticator() *None {
	return &None{}
}

func (n *None) Authenticate(w http.ResponseWriter, r *http.Request) {
}

func (n *None) Authorized(w http.ResponseWriter, r *http.Request, path string) (string, string, bool) {
	return "-", "-", true
}

func (n *None) Logout(w http.ResponseWriter, r *http.Request) {
}

func (n *None) Session(r *http.Request) (*session, error) {
	return nil, nil
}

func (n *None) Sweep() {
}
