package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

const (
	LoginCookie   = "uhppoted-httpd-login"
	SessionCookie = "uhppoted-httpd-session"
)

type Basic struct {
	auth         auth.IAuthenticate
	cookieMaxAge int
	resources    []resource
	sync.RWMutex
}

type resource struct {
	Path       *regexp.Regexp `json:"path"`
	Authorised *regexp.Regexp `json:"authorised"`
}

func NewBasic(auth auth.IAuthenticate, file string, cookieMaxAge int) (*Basic, error) {
	a := Basic{
		auth:         auth,
		cookieMaxAge: cookieMaxAge,
	}

	if err := a.load(file); err != nil {
		return nil, err
	}

	a.watch(file)

	return &a, nil
}

func (b *Basic) Preauthenticate() (*http.Cookie, error) {
	token, err := b.auth.Preauthenticate()
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

// NTS: the uhppoted-httpd-login cookie is a single use expiring cookie
//
//	intended to (eventually) support opaque login credentials
func (b *Basic) Authenticate(uid, pwd string, cookie *http.Cookie) (*http.Cookie, error) {
	if cookie == nil {
		return nil, fmt.Errorf("Invalid login cookie")
	}

	if err := b.auth.Verify(auth.Login, cookie.Value); err != nil {
		return nil, err
	}

	b.auth.Invalidate(auth.Login, cookie.Value)

	if token, err := b.auth.Authenticate(uid, pwd); err != nil {
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

func (b *Basic) Verify(uid, pwd string) error {
	return b.auth.Validate(uid, pwd)
}

func (b *Basic) Logout(cookie *http.Cookie) {
	if err := b.auth.Invalidate(auth.Session, cookie.Value); err != nil {
		warn(err)
	}
}

func (b *Basic) Authorised(uid, role, path string) error {
	b.RLock()
	defer b.RUnlock()

	for _, r := range b.resources {
		if r.Path.MatchString(path) && r.Authorised.MatchString(role) {
			return nil
		}
	}

	return fmt.Errorf("%v not authorized for %s", uid, path)
}

func (b *Basic) load(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	return b.deserialize(f)
}

func (b *Basic) deserialize(r io.Reader) error {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	serializable := struct {
		Resources []resource `json:"resources"`
	}{
		Resources: []resource{},
	}

	if err := json.Unmarshal(bytes, &serializable); err != nil {
		return err
	}

	b.Lock()
	b.resources = serializable.Resources
	b.Unlock()

	return nil
}

// NOTE: interim file watcher implementation pending fsnotify in Go v?.?
//
//	(https://github.com/fsnotify/fsnotify requires workarounds for
//	 files updated atomically by renaming)
func (b *Basic) watch(filepath string) {
	go func() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			log.Printf("ERROR Failed to get file information for '%s': %v", filepath, err)
			return
		}

		lastModified := finfo.ModTime()
		logged := false
		for {
			time.Sleep(2500 * time.Millisecond)
			finfo, err := os.Stat(filepath)
			if err != nil {
				if !logged {
					log.Printf("ERROR Failed to get file information for '%s': %v", filepath, err)
					logged = true
				}

				continue
			}

			logged = false
			if finfo.ModTime() != lastModified {
				log.Printf("INFO  Reloading information from %s\n", filepath)

				err := b.load(filepath)
				if err != nil {
					log.Printf("ERROR Failed to reload information from %s: %v", filepath, err)
					continue
				}

				log.Printf("INFO  Updated auth DB from %s", filepath)
				lastModified = finfo.ModTime()
			}
		}
	}()
}

func (r *resource) UnmarshalJSON(bytes []byte) error {
	x := struct {
		Path       string `json:"path"`
		Authorised string `json:"authorised"`
	}{}

	err := json.Unmarshal(bytes, &x)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(x.Path, "^") {
		x.Path = "^" + x.Path
	}

	if !strings.HasSuffix(x.Path, "$") {
		x.Path = x.Path + "$"
	}

	path, err := regexp.Compile(x.Path)
	if err != nil {
		return err
	}

	authorised, err := regexp.Compile(x.Authorised)
	if err != nil {
		return err
	}

	r.Path = path
	r.Authorised = authorised

	return nil
}
