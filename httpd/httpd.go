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

	"github.com/uhppoted/uhppoted-httpd/auth"
)

type HTTPD struct {
	Dir          string
	AuthProvider auth.IAuth
	CookieMaxAge int
}

type dispatcher struct {
	root         string
	fs           http.Handler
	auth         auth.IAuth
	cookieMaxAge int
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
	if cookie, err := r.Cookie("uhppoted-httpd-auth"); err == nil {
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
		if cookie, err := r.Cookie("uhppoted-httpd-auth"); err == nil {
			if err := d.auth.Authorized(cookie.Value, path); err != nil {
				info(err.Error())
			} else {
				return true
			}
		}

		return false
	}

	return true
}

func (d *dispatcher) user(r *http.Request) string {
	if cookie, err := r.Cookie("uhppoted-httpd-auth"); err == nil {
		if uid, err := d.auth.User(cookie.Value); err != nil {
			info(err.Error())
		} else {
			return uid
		}
	}

	return ""
}

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
}

func (d *dispatcher) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/unauthorized.html", http.StatusFound)
}

func (d *dispatcher) authenticate(w http.ResponseWriter, r *http.Request) {
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

	token, err := d.auth.Authorize(uid, pwd)
	if err != nil {
		d.unauthorized(w, r)
		return
	}

	cookie := http.Cookie{
		Name:     "uhppoted-httpd-auth",
		Value:    token,
		Path:     "/",
		MaxAge:   d.cookieMaxAge * int(time.Hour.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func authorize(header []string) error {
	if len(header) == 0 {
		return fmt.Errorf("Empty 'Authorization' header")
	}

	return nil
}

type httpdFileSystem struct {
	http.FileSystem
}

func (fs httpdFileSystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return nil, os.ErrPermission
		}
	}

	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	return httpdFile{file}, err
}

type httpdFile struct {
	http.File
}

func (f httpdFile) Readdir(n int) (fis []os.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}

	return
}

func debug(message string) {
	log.Printf("%-5s %s", "DEBUG", message)
}

func info(message string) {
	log.Printf("%-5s %s", "INFO", message)
}

func warn(message string, err error) {
	if err == nil {
		log.Printf("%-5s %s", "WARN", message)
	} else {
		log.Printf("%-5s %s", "WARN", message)
		log.Printf("%-5s %v", "", err)
	}
}
