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

func (n *None) Preauthenticate() (*http.Cookie, error) {
	return nil, nil
}

func (n *None) Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error) {
	return nil, nil
}

func (n *None) Authenticated(r *http.Request) (string, string, bool) {
	return "-", "-", true
}

func (n *None) Authorised(uid, role, path string) error {
	return nil
}

func (n *None) Verify(uid, pwd string) error {
	return nil
}

func (n *None) SetPassword(uid, pwd, role string) error {
	return nil
}

func (n *None) Logout(w http.ResponseWriter, r *http.Request) {
}

func (n *None) Session(r *http.Request) (*session, error) {
	return nil, nil
}

func (n *None) Sweep() {
}
