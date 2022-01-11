package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/auth"
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
		MaxAge:   b.cookieMaxAge * int(time.Hour.Seconds()),
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

	if token, err := b.auth.Authorize(uid, pwd, sessionId); err != nil {
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

			User: uid,
		}

		return &cookie, nil
	}
}

func (b *Basic) Authenticated(r *http.Request) (string, string, bool) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		warn(fmt.Errorf("No JWT cookie in request"))
		return "", "", false
	}

	uid, role, err := b.auth.Authenticated(cookie.Value)
	if err != nil {
		warn(err)
		return "", "", false
	}

	session, err := b.session(r)
	if err != nil {
		warn(err)
		return "", "", false
	}

	if session == nil {
		warn(fmt.Errorf("No extant session for request"))
		return "", "", false
	}

	session.touched = time.Now()

	return uid, role, true
}

func (b *Basic) Authorised(uid, role, path string) error {
	return b.auth.AuthorisedX(uid, role, path)
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

func (b *Basic) Logout(w http.ResponseWriter, r *http.Request) {
	if s, _ := b.session(r); s != nil {
		delete(b.sessions, s.id)
	}

	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (b *Basic) Session(r *http.Request) (*session, error) {
	return b.session(r)
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

func (b *Basic) authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		warn(err)
		return false
	}

	if err = b.auth.Verify(auth.Session, cookie.Value); err != nil {
		warn(err)
		return false
	}

	return true
}

func (b *Basic) authorized(r *http.Request, path string) (string, string, bool) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		warn(fmt.Errorf("No JWT cookie in request"))
		return "", "", false
	}

	uid, role, err := b.auth.Authorized(cookie.Value, path)
	if err != nil {
		warn(err)
		return "", "", false
	}

	session, err := b.session(r)
	if err != nil {
		warn(err)
		return "", "", false
	}

	if session == nil {
		warn(fmt.Errorf("No extant session for request"))
		return "", "", false
	}

	session.touched = time.Now()

	return uid, role, true
}

func (b *Basic) validateLoginCookie(cookie *http.Cookie) error {
	if cookie == nil {
		return fmt.Errorf("Invalid login cookie")
	}

	if err := b.auth.Verify(auth.Login, cookie.Value); err != nil {
		return err
	}

	lid, err := b.auth.GetLoginId(cookie.Value)
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

func (b *Basic) session(r *http.Request) (*session, error) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return nil, err
	}

	sid, err := b.auth.GetSessionId(cookie.Value)
	if err != nil {
		return nil, err
	}

	if sid == nil {
		return nil, fmt.Errorf("Invalid session ID (%v)", sid)
	}

	s, ok := b.sessions[*sid]
	if !ok {
		return nil, fmt.Errorf("No extant session for session ID '%v'", *sid)
	}

	return s, nil
}
