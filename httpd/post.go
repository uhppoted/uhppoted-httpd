package httpd

import (
	"net/http"
	"regexp"

	"github.com/uhppoted/uhppoted-httpd/httpd/cardholders"
	"github.com/uhppoted/uhppoted-httpd/httpd/system"
)

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	uid, role, ok := d.authorized(w, r, path)
	if !ok {
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	auth, err := NewAuthorizator(uid, role, d.grule.cards)
	if err != nil {
		http.Error(w, "Error executing request", http.StatusInternalServerError)
	}

	if match, err := regexp.MatchString(`/system`, path); err == nil && match {
		system.Post(w, r, d.timeout, d.db, auth)
		return
	}

	if match, err := regexp.MatchString(`/cardholders`, path); err == nil && match {
		cardholders.Post(w, r, d.timeout, d.db, auth)
		return
	}

	http.Error(w, "API not implemented", http.StatusNotImplemented)
}
