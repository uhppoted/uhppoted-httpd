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
	logins       map[uuid.UUID]*login
	sessions     map[uuid.UUID]*session
	stale        time.Duration
}

func NewBasicAuthenticator(auth auth.IAuth, cookieMaxAge int, stale time.Duration) *Basic {
	a := Basic{
		auth:         auth,
		cookieMaxAge: cookieMaxAge,
		logins:       map[uuid.UUID]*login{},
		sessions:     map[uuid.UUID]*session{},
		stale:        stale,
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

	b.logins[loginId] = &login{
		id:      loginId,
		touched: time.Now(),
	}

	return &cookie, nil
}

// NOTE TO SELF: the uhppoted-httpd-login cookie is a single use expiring cookie
//               intended to (eventually) support opaque login credentials
func (b *Basic) Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error) {
	if err := b.validateLoginCookie(cookie); err != nil {
		return nil, err
	}

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

		b.sessions[sessionId] = &session{
			id:      sessionId,
			touched: time.Now(),
		}

		return &cookie, nil
	}
}

func (b *Basic) Authenticated(cookie *http.Cookie) (string, string, *http.Cookie, error) {
	uid, role, sid, token, err := b.auth.Authenticated(cookie.Value)
	if err != nil {
		return "", "", nil, err
	}

	if sid == nil {
		return "", "", nil, fmt.Errorf("Invalid session ID (%v)", sid)
	} else if session, ok := b.sessions[*sid]; !ok {
		return "", "", nil, fmt.Errorf(">>>>> DEBUG/1 No extant session for session ID '%v'", *sid)
	} else if session == nil {
		return "", "", nil, fmt.Errorf("No extant session for request")
	} else {
		session.touched = time.Now()
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
	if err := b.auth.Invalidate(cookie.Value); err != nil {
		warn(err)
	}

	if s, _ := b.session(cookie); s != nil {
		delete(b.sessions, s.id)
	}
}

func (b *Basic) Sweep() {
	cutoff := time.Now().Add(-b.stale)

	for k, v := range b.logins {
		if v.touched.Before(cutoff) {
			warn(fmt.Errorf("Removing stale login ID for %v", k))
			delete(b.logins, k)
		}
	}

	for k, v := range b.sessions {
		if v.touched.Before(cutoff) {
			warn(fmt.Errorf("Removing stale session ID for %v", k))
			delete(b.sessions, k)
		}
	}
}

func (b *Basic) validateLoginCookie(cookie *http.Cookie) error {
	if cookie == nil {
		return fmt.Errorf("Invalid login cookie")
	}

	lid, err := b.auth.Verify(auth.Login, cookie.Value)
	if err != nil {
		return err
	}

	if lid == nil {
		return fmt.Errorf("Invalid login ID (%v)", lid)
	}

	if _, ok := b.logins[*lid]; !ok {
		return fmt.Errorf("No extant login for login ID '%v'", *lid)
	}

	delete(b.logins, *lid)

	return nil
}

func (b *Basic) session(cookie *http.Cookie) (*session, error) {
	if cookie == nil {
		return nil, fmt.Errorf("Invalid session cookie")
	}

	sid, err := b.auth.Verify(auth.Session, cookie.Value)
	if err != nil {
		return nil, err
	}

	if sid == nil {
		return nil, fmt.Errorf("Invalid session ID (%v)", sid)
	}

	s, ok := b.sessions[*sid]
	if !ok {
		return nil, fmt.Errorf(">>>>> DEBUG/2 No extant session for session ID '%v'", *sid)
	}

	return s, nil
}
