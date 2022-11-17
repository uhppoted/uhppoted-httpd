package httpd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	authorizator "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
	"github.com/uhppoted/uhppoted-httpd/httpd/post"
	"github.com/uhppoted/uhppoted-httpd/httpd/users"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	// ... allow unauthenticated access to /authenticate and /logout
	if path == "/authenticate" {
		post.Login(w, r, d.auth)
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	// ... require auth for everything else
	uid, role, authenticated := d.authenticated(r, w)
	if !authenticated {
		d.unauthenticated(r, w)
		return
	}

	if ok := d.authorised(uid, role, path); !ok {
		d.unauthorised(r, w)
		return
	}

	switch path {
	case "/password":
		d.exec(w, r, func(m map[string]interface{}) (interface{}, error) {
			return users.Password(m, role, d.auth)
		})

	case
		"/interfaces",
		"/controllers",
		"/doors",
		"/cards",
		"/groups",
		"/users":
		if handler := d.vtable(path); handler == nil || handler.post == nil {
			warnf("HTTPD", "No vtable entry for %v", path)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else if d.mode == types.Monitor {
			warnf("HTTPD", "POST request in 'monitor' mode")
			http.Error(w, "Configuration changes are disabled in monitor-only mode", http.StatusBadRequest)
		} else {
			d.exec(w, r, func(m map[string]interface{}) (interface{}, error) {
				return handler.post(uid, role, m)
			})
		}

	case "/synchronize/ACL":
		if d.mode == types.Monitor {
			http.Error(w, "Synchronize ACL disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			post.SynchronizeACL(d.context, w, r, d.timeout)
		}

	case "/synchronize/datetime":
		if d.mode == types.Monitor {
			http.Error(w, "Synchronize date/time disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			d.synchronize(w, r, system.SynchronizeDateTime)
		}

	case "/synchronize/doors":
		if d.mode == types.Monitor {
			http.Error(w, "Synchronize doors disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			d.synchronize(w, r, system.SynchronizeDoors)
		}

	case "/otp":
		users.VerifyOTP(uid, role, w, r, d.auth)

	default:
		http.Error(w, "API not implemented", http.StatusNotImplemented)
	}
}

func (d *dispatcher) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(cookies.SessionCookie); err == nil {
		d.auth.Logout(cookie)
	}

	cookies.Clear(w, cookies.SessionCookie, cookies.OTPCookie)
	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (d *dispatcher) exec(w http.ResponseWriter, r *http.Request, f func(map[string]interface{}) (interface{}, error)) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		defer close(ch)

		acceptsGzip := false
		contentType := ""

		for k, h := range r.Header {
			if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
				for _, v := range h {
					contentType = strings.TrimSpace(strings.ToLower(v))
				}
			}

			if strings.TrimSpace(strings.ToLower(k)) == "accept-encoding" {
				for _, v := range h {
					if strings.Contains(strings.TrimSpace(strings.ToLower(v)), "gzip") {
						acceptsGzip = true
					}
				}
			}
		}

		body := map[string]interface{}{}

		switch contentType {
		case "application/x-www-form-urlencoded":
			if err := r.ParseForm(); err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Error reading request", http.StatusInternalServerError)
				return
			}

			for k, v := range r.Form {
				body[k] = v
			}

		case "application/json":
			blob, err := ioutil.ReadAll(r.Body)
			if err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Error reading request", http.StatusInternalServerError)
				return
			}

			if err := json.Unmarshal(blob, &body); err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, fmt.Sprintf("Invalid request content-type (%v)", contentType), http.StatusBadRequest)
			return
		}

		response, err := f(body)
		if err != nil && errors.Is(err, authorizator.Unauthorised) {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := json.Marshal(response)
		if err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, "Internal error generating response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if acceptsGzip && len(b) > GZIP_MINIMUM {
			w.Header().Set("Content-Encoding", "gzip")

			gz := gzip.NewWriter(w)
			gz.Write(b)
			gz.Close()
		} else {
			w.Write(b)
		}
	}()

	select {
	case <-ctx.Done():
		warnf("HTTPD", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return

	case <-ch:
	}
}

func (d *dispatcher) synchronizeDateTime(w http.ResponseWriter, r *http.Request) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		if err := system.SynchronizeDateTime(); err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warnf("HTTPD", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)

	case <-ch:
	}
}

func (d *dispatcher) synchronize(w http.ResponseWriter, r *http.Request, f func() error) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		if err := f(); err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warnf("HTTPD", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)

	case <-ch:
	}
}
