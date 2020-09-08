package httpd

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/cardholders"
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

	case "/cardholders":
		cardholders.Update(d.db, w, r, d.timeout)

	default:
		http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
	}
}
