package httpd

import (
	"net/http"
)

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	if !d.authorised(r, path) {
		if !d.authenticated(r) {
			http.Redirect(w, r, "/login.html", http.StatusFound)
		} else if s := d.session(r); s == nil {
			http.Redirect(w, r, "/login.html", http.StatusFound)
		} else {
			d.unauthorized(w, r)
		}

		return
	}

	switch path {
	case "/logout":
		d.logout(w, r)

	default:
		http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
	}
}
