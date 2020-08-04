package httpd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

const (
	JWTCookie = "uhppoted-httpd-jwt"
)

type HTTPD struct {
	Dir          string
	AuthProvider auth.IAuth
	CookieMaxAge int
}

type session struct {
	id   uuid.UUID
	user string
}

type dispatcher struct {
	root         string
	fs           http.Handler
	auth         auth.IAuth
	cookieMaxAge int
	sessions     map[uuid.UUID]*session
}

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	d := dispatcher{
		root:         h.Dir,
		fs:           http.FileServer(fs),
		auth:         h.AuthProvider,
		cookieMaxAge: h.CookieMaxAge,
		sessions:     map[uuid.UUID]*session{},
	}

	srv := http.Server{
		Addr: ":8080",
	}

	shutdown := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("WARN  HTTP server shutdown error: %v", err)
		}

		close(shutdown)
	}()

	http.Handle("/", &d)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ERROR: %v", err)
	}

	<-shutdown
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	debug(fmt.Sprintf("%v", r.URL))

	switch strings.ToUpper(r.Method) {
	case http.MethodGet:
		d.get(w, r)
	case http.MethodPost:
		d.post(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}
}

func (d *dispatcher) authenticated(r *http.Request) bool {
	if cookie, err := r.Cookie(JWTCookie); err == nil {
		if err := d.auth.Verify(cookie.Value); err != nil {
			info(err.Error())
		} else {
			return true
		}
	}

	return false
}

func (d *dispatcher) authorised(r *http.Request, path string) bool {
	if path == "/login.html" {
		return true
	}

	if path == "/unauthorized.html" {
		return true
	}

	if strings.HasSuffix(path, ".html") {
		cookie, err := r.Cookie(JWTCookie)
		if err != nil {
			warn(fmt.Errorf("No JWT cookie in request"))
			return false
		}

		if err := d.auth.Authorized(cookie.Value, path); err != nil {
			warn(err)
			return false
		}

		session, err := d.session(r)
		if err != nil {
			warn(err)
			return false
		}

		if session == nil {
			warn(fmt.Errorf("No extant session for request"))
			return false
		}
	}

	return true
}

func (d *dispatcher) session(r *http.Request) (*session, error) {
	cookie, err := r.Cookie(JWTCookie)
	if err != nil {
		return nil, err
	}

	sid, err := d.auth.Session(cookie.Value)
	if err != nil {
		return nil, err
	}

	if sid == nil {
		return nil, fmt.Errorf("Invalid session ID (%v)", sid)
	}

	s, ok := d.sessions[*sid]
	if !ok {
		return nil, fmt.Errorf("No extant session for session ID '%v'", *sid)
	}

	return s, nil
}

func (d *dispatcher) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/unauthorized.html", http.StatusFound)
}

func (d *dispatcher) authenticate(w http.ResponseWriter, r *http.Request) {
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
		Name:     JWTCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   d.cookieMaxAge * int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	}

	d.sessions[sessionId] = &session{
		id:   sessionId,
		user: uid,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (d *dispatcher) logout(w http.ResponseWriter, r *http.Request) {
	if s, _ := d.session(r); s != nil {
		delete(d.sessions, s.id)
	}

	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func authorize(header []string) error {
	if len(header) == 0 {
		return fmt.Errorf("Empty 'Authorization' header")
	}

	return nil
}

func debug(message string) {
	log.Printf("%-5s %s", "DEBUG", message)
}

func info(message string) {
	log.Printf("%-5s %s", "INFO", message)
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
