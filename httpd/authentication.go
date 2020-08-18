package httpd

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

func (d *dispatcher) unauthenticated(w http.ResponseWriter, r *http.Request) {
	loginId := uuid.New()
	token, err := d.auth.Preauthenticate(loginId)
	if err != nil {
		http.Error(w, "Invalid login token", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     LoginCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   d.cookieMaxAge * int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	}

	d.logins[loginId] = &login{
		id:      loginId,
		touched: time.Now(),
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func (d *dispatcher) authenticate(w http.ResponseWriter, r *http.Request) {
	if err := d.validateLoginCookie(r); err != nil {
		warn(err)
		d.unauthenticated(w, r)
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
		r.ParseForm()
		if v, ok := r.Form["uid"]; ok && len(v) > 0 {
			uid = v[0]
		}

		if v, ok := r.Form["pwd"]; ok && len(v) > 0 {
			pwd = v[0]
		}

	case "application/json":
		blob, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		body := struct {
			UserId   string `json:"uid"`
			Password string `json:"pwd"`
		}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		uid = body.UserId
		pwd = body.Password
	}

	token, err := d.auth.Authorize(uid, pwd, sessionId)
	if err != nil {
		d.unauthorized(w, r)
		return
	}

	cookie := http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   d.cookieMaxAge * int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	}

	d.sessions[sessionId] = &session{
		id:      sessionId,
		user:    uid,
		touched: time.Now(),
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (d *dispatcher) authenticated(r *http.Request) bool {
	if cookie, err := r.Cookie(SessionCookie); err == nil {
		if err := d.auth.Verify(auth.Session, cookie.Value); err != nil {
			info(err.Error())
		} else {
			return true
		}
	}

	return false
}

func (d *dispatcher) validateLoginCookie(r *http.Request) error {
	cookie, err := r.Cookie(LoginCookie)
	if err != nil {
		return err
	}

	if err := d.auth.Verify(auth.Login, cookie.Value); err != nil {
		return err
	}

	lid, err := d.auth.GetLoginId(cookie.Value)
	if err != nil {
		return err
	}

	if lid == nil {
		return fmt.Errorf("Invalid login ID (%v)", lid)
	}

	_, ok := d.logins[*lid]
	if !ok {
		return fmt.Errorf("No extant login for login ID '%v'", *lid)
	}

	delete(d.logins, *lid)

	return nil
}
