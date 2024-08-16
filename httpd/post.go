package httpd

import (
	"context"
	"net/http"

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

	// ... allow unauthenticated access to /authenticate, /logout and /setup
	if path == "/authenticate" {
		post.Login(w, r, d.auth)
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	if path == "/setup" && !d.noSetup {
		d.setup(w, r)
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

	case "/password":
		d.exec2(w, r, func() (any, error) {
			return users.Password(uid, role, w, r, d.auth)
		})

	case "/otp":
		d.exec2(w, r, func() (any, error) {
			return users.VerifyOTP(uid, role, w, r, d.auth)
		})

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

func (d *dispatcher) setup(w http.ResponseWriter, r *http.Request) {
	_, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	// ... disallow logged in users
	uid, role, authenticated := d.authenticated(r, w)
	if uid != "" || role != "" || authenticated {
		http.Error(w, "Not allowed for logged in user", http.StatusForbidden)
		return
	}

	users.Setup(w, r, d.auth)
}
