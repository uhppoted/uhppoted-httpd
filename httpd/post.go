package httpd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
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
		d.login(w, r)
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
			warn("", fmt.Errorf("No vtable entry for %v", path))
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else if d.mode == Monitor {
			warn("", fmt.Errorf("POST request in 'monitor' mode"))
			http.Error(w, "Configuration changes are disabled in monitor-only mode", http.StatusBadRequest)
		} else {
			d.exec(w, r, func(m map[string]interface{}) (interface{}, error) {
				return handler.post(uid, role, m)
			})
		}

	case "/synchronize/ACL":
		if d.mode == Monitor {
			http.Error(w, "Synchronize ACL disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			d.synchronize(w, r, system.SynchronizeACL)
		}

	case "/synchronize/datetime":
		if d.mode == Monitor {
			http.Error(w, "Synchronize date/time disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			d.synchronize(w, r, system.SynchronizeDateTime)
		}

	case "/synchronize/doors":
		if d.mode == Monitor {
			http.Error(w, "Synchronize doors disabled in 'monitor' mode", http.StatusBadRequest)
		} else {
			d.synchronize(w, r, system.SynchronizeDoors)
		}

	default:
		http.Error(w, "API not implemented", http.StatusNotImplemented)
	}
}

func (d *dispatcher) login(w http.ResponseWriter, r *http.Request) {
	var contentType string
	var uid string
	var pwd string

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
			warn("", err)
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		body := struct {
			UserId   string `json:"uid"`
			Password string `json:"pwd"`
		}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			warn("", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		uid = body.UserId
		pwd = body.Password
	}

	loginCookie, err := r.Cookie(auth.LoginCookie)
	if err != nil {
		warn("", err)
		d.unauthenticated(r, w)
		return
	}

	if loginCookie == nil {
		warn("", fmt.Errorf("Missing login cookie"))
		http.Error(w, "Missing login cookie", http.StatusBadRequest)
		return
	}

	sessionCookie, err := d.auth.Authenticate(uid, pwd, loginCookie)
	if err != nil {
		warn("", err)
		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	if sessionCookie != nil {
		http.SetCookie(w, sessionCookie)
	}

	clear(auth.LoginCookie, w)
}

func (d *dispatcher) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookie); err == nil {
		d.auth.Logout(cookie)
	}

	clear(auth.SessionCookie, w)

	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func (d *dispatcher) exec(w http.ResponseWriter, r *http.Request, f func(map[string]interface{}) (interface{}, error)) {
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
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
				ch <- &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid reading request"),
					Detail: err,
				}
				return
			}

			for k, v := range r.Form {
				body[k] = v
			}

		case "application/json":
			blob, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ch <- &types.HttpdError{
					Status: http.StatusInternalServerError,
					Err:    fmt.Errorf("Invalid reading request"),
					Detail: err,
				}
				return
			}

			if err := json.Unmarshal(blob, &body); err != nil {
				ch <- &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid request body"),
					Detail: fmt.Errorf("Error unmarshalling request (%s): %w", string(blob), err),
				}
				return
			}

		default:
			ch <- &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid request"),
				Detail: fmt.Errorf("Invalid request content-type (%v)", contentType),
			}
			return
		}

		response, err := f(body)
		if err != nil {
			ch <- err
			return
		}

		b, err := json.Marshal(response)
		if err != nil {
			ch <- &types.HttpdError{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("Internal error generating response"),
				Detail: fmt.Errorf("Error marshalling response: %w", err),
			}
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

		ch <- nil
	}()

	select {
	case <-ctx.Done():
		warn("", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return

	case err := <-ch:
		if err != nil {
			warn("", err)

			switch e := err.(type) {
			case *types.HttpdError:
				http.Error(w, e.Error(), e.Status)

			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}

			return
		}
	}
}

func (d *dispatcher) synchronizeACL(w http.ResponseWriter, r *http.Request) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		if err := system.SynchronizeACL(); err != nil {
			warn("", err)

			switch e := err.(type) {
			case *types.HttpdError:
				http.Error(w, e.Error(), e.Status)

			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warn("", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)

	case <-ch:
	}
}

func (d *dispatcher) synchronizeDateTime(w http.ResponseWriter, r *http.Request) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		if err := system.SynchronizeDateTime(); err != nil {
			warn("", err)

			switch e := err.(type) {
			case *types.HttpdError:
				http.Error(w, e.Error(), e.Status)

			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warn("", ctx.Err())
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
			warn("", err)

			switch e := err.(type) {
			case *types.HttpdError:
				http.Error(w, e.Error(), e.Status)

			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warn("", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)

	case <-ch:
	}
}
