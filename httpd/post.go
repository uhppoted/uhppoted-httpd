package httpd

import (
	"net/http"
	"regexp"

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

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	if match, err := regexp.MatchString(`/cardholders/\S+`, path); err == nil && match {
		cardholders.Update(d.db, w, r, d.timeout)
		return
	}

	http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
}
