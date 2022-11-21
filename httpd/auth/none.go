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

func (n *None) Authenticated(cookie *http.Cookie) (string, string, *http.Cookie, error) {
	return "-", "-", nil, nil
}

func (n *None) Authorised(uid, role, path string) error {
	return nil
}

func (n *None) Verify(uid, pwd string) error {
	return nil
}

func (n *None) VerifyAuthHeader(uid string, header string) error {
	return nil
}

func (n *None) Options(uid string) options {
	return options{}
}

func (n *None) Logout(cookie *http.Cookie) {
}
