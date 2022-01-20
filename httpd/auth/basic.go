package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

const (
	LoginCookie   = "uhppoted-httpd-login"
	SessionCookie = "uhppoted-httpd-session"
)

type Basic struct {
	auth         auth.IAuth
	cookieMaxAge int
}

func NewBasicAuthenticator(auth auth.IAuth, cookieMaxAge int) *Basic {
	a := Basic{
		auth:         auth,
		cookieMaxAge: cookieMaxAge,
	}

	return &a
}

func (b *Basic) Preauthenticate() (*http.Cookie, error) {
	loginId := uuid.New()

	token, err := b.auth.Preauthenticate(loginId)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:     LoginCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   300 * int(time.Hour.Seconds()), // 5 minutes
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	}

	return &cookie, nil
}

// NOTE TO SELF: the uhppoted-httpd-login cookie is a single use expiring cookie
//               intended to (eventually) support opaque login credentials
func (b *Basic) Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error) {
	if cookie == nil {
		return nil, fmt.Errorf("Invalid login cookie")
	}

	if err := b.auth.Verify(auth.Login, cookie.Value); err != nil {
		return nil, err
	}

	b.auth.Invalidate(auth.Login, cookie.Value)

	var sessionId = uuid.New()

	if token, err := b.auth.Authenticate(uid, pwd, sessionId); err != nil {
		return nil, err
	} else {
		cookie := http.Cookie{
			Name:     SessionCookie,
			Value:    token,
			Path:     "/",
			MaxAge:   b.cookieMaxAge * int(time.Hour.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			//	Secure:   true,
		}

		return &cookie, nil
	}
}

func (b *Basic) Authenticated(cookie *http.Cookie) (string, string, *http.Cookie, error) {
	uid, role, token, err := b.auth.Authenticated(cookie.Value)
	if err != nil {
		return "", "", nil, err
	}

	if token != "" {
		cookie := http.Cookie{
			Name:     SessionCookie,
			Value:    token,
			Path:     "/",
			MaxAge:   b.cookieMaxAge * int(time.Hour.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			//	Secure:   true,
		}

		return uid, role, &cookie, nil
	}

	return uid, role, nil, nil
}

func (b *Basic) Authorised(uid, role, path string) error {
	return b.auth.Authorised(uid, role, path)
}

func (b *Basic) Verify(uid, pwd string) error {
	return b.auth.Validate(uid, pwd)
}

func (b *Basic) SetPassword(uid, pwd, role string) error {
	if err := b.auth.Store(uid, pwd, role); err != nil {
		return err
	}

	return b.auth.Save()
}

func (b *Basic) Logout(cookie *http.Cookie) {
	if err := b.auth.Invalidate(auth.Session, cookie.Value); err != nil {
		warn(err)
	}
}
