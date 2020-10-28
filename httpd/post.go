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

	uid, ok := d.authorized(w, r, path)
	if !ok {
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	if match, err := regexp.MatchString(`/cardholders(?:/.*)?`, path); err == nil && match {
		cardholders.Post(uid, d.db, w, r, d.timeout)
		return
	}

	http.Error(w, "API not implemented", http.StatusNotImplemented)
}
