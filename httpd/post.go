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

	if match, err := regexp.MatchString(`/system`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, d.grule.system)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		system.Post(w, r, d.timeout, auth)
		return
	}

	if match, err := regexp.MatchString(`/cardholders`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, d.grule.cards)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		cardholders.Post(w, r, d.timeout, auth)
		return
	}

	http.Error(w, "API not implemented", http.StatusNotImplemented)
}
