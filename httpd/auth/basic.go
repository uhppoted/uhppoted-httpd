package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	urls         map[string]struct{}
}

func NewBasicAuthenticator(auth auth.IAuth, cookieMaxAge int, stale time.Duration, urls []string) *Basic {
	a := Basic{
		auth:         auth,
		cookieMaxAge: cookieMaxAge,
		logins:       map[uuid.UUID]*login{},
		sessions:     map[uuid.UUID]*session{},
		stale:        stale,
		urls:         map[string]struct{}{},
	}

	for _, u := range urls {
		a.urls[u] = struct{}{}
	}

	return &a
}

func (b *Basic) Verify(uid, pwd string, r *http.Request) error {
	return fmt.Errorf("NOT IMPLEMENTED")
	// if err := b.validateLoginCookie(r); err != nil {
	// 	warn(err)
	// 	b.unauthenticated(w, r)
	// 	return
	// }

	// var sessionId = uuid.New()
	// var uid string
	// var pwd string
	// var contentType string

	// for k, h := range r.Header {
	// 	if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
	// 		for _, v := range h {
	// 			contentType = strings.TrimSpace(strings.ToLower(v))
	// 		}
	// 	}
	// }

	// switch contentType {
	// case "application/x-www-form-urlencoded":
	// 	uid = r.FormValue("uid")
	// 	pwd = r.FormValue("pwd")

	// case "application/json":
	// 	blob, err := ioutil.ReadAll(r.Body)
	// 	if err != nil {
	// 		warn(err)
	// 		http.Error(w, "Error reading request", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	body := struct {
	// 		UserId   string `json:"uid"`
	// 		Password string `json:"pwd"`
	// 	}{}

	// 	if err := json.Unmarshal(blob, &body); err != nil {
	// 		warn(err)
	// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
	// 		return
	// 	}

	// 	uid = body.UserId
	// 	pwd = body.Password
	// }

	// token, err := b.auth.Authorize(uid, pwd, sessionId)
	// if err != nil {
	// 	warn(err)
	// 	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	// 	return
	// }

	// cookie := http.Cookie{
	// 	Name:     SessionCookie,
	// 	Value:    token,
	// 	Path:     "/",
	// 	MaxAge:   b.cookieMaxAge * int(time.Hour.Seconds()),
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteStrictMode,
	// 	//	Secure:   true,
	// }

	// b.sessions[sessionId] = &session{
	// 	id:      sessionId,
	// 	touched: time.Now(),

	// 	User: uid,
	// }

	// http.SetCookie(w, &cookie)
}

// NTS: the uhppoted-httpd-login cookie is a single use expiring cookie intended to
//      (eventually) support opaque login credentials
func (b *Basic) Authenticate(w http.ResponseWriter, r *http.Request) {
	// HEAD request refreshes uhppoted-httpd-login cookie
	if strings.ToUpper(r.Method) == http.MethodHead {
		if err := b.setLoginCookie(w); err != nil {
			warn(err)
			return
		}

		return
	}

	// POST request validates uhppoted-httpd-login cookie and credentials
	if err := b.validateLoginCookie(r); err != nil {
		warn(err)
		b.unauthenticated(w, r)
		return
	}

	var sessionId = uuid.New()
	var uid string
	var pwd string
	var contentType string

	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
			for _, v := range h {
				contentType = strings.TrimSpace(strings.ToLower(v))
			}
		}
	}

	switch contentType {
	case "application/x-www-form-urlencoded":
		uid = r.FormValue("uid")
		pwd = r.FormValue("pwd")

	case "application/json":
		blob, err := ioutil.ReadAll(r.Body)
		if err != nil {
			warn(err)
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		body := struct {
			UserId   string `json:"uid"`
			Password string `json:"pwd"`
		}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			warn(err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		uid = body.UserId
		pwd = body.Password
	}

	token, err := b.auth.Authorize(uid, pwd, sessionId)
	if err != nil {
		warn(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

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

	http.SetCookie(w, &cookie)
}

func (b *Basic) Authorized(w http.ResponseWriter, r *http.Request, path string) (string, string, bool) {
	uid, role, ok := b.authorized(r, path)
	if !ok {
		if !b.authenticated(r) {
			b.unauthenticated(w, r)
		} else if s, err := b.session(r); err != nil || s == nil {
			b.unauthenticated(w, r)
		} else {
			b.unauthorized(w, r)
		}

		return "", "", false
	}

	return uid, role, true
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
	if path == "/login.html" {
		return "", "", true
	}

	if path == "/unauthorized.html" {
		return "", "", true
	}

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

func (b *Basic) unauthenticated(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func (b *Basic) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/unauthorized.html", http.StatusFound)
}

func (b *Basic) validateLoginCookie(r *http.Request) error {
	cookie, err := r.Cookie(LoginCookie)
	if err != nil {
		return err
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
		if _, ok := b.urls[r.URL.Path]; !ok {
			return fmt.Errorf("No extant login for login ID '%v'", *lid)
		}
	}

	delete(b.logins, *lid)

	return nil
}

func (b *Basic) setLoginCookie(w http.ResponseWriter) error {
	loginId := uuid.New()
	token, err := b.auth.Preauthenticate(loginId)
	if err != nil {
		http.Error(w, "Invalid login token", http.StatusInternalServerError)
		return nil
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

	http.SetCookie(w, &cookie)

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
