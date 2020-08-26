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

	if !d.authorized(w, r, path) {
		return
	}

	switch path {
	case "/logout":
		d.logout(w, r)

	default:
		http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
	}
}
