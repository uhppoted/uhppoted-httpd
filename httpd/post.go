package httpd

import (
	"net/http"
	"regexp"

	"github.com/uhppoted/uhppoted-httpd/httpd/cards"
	"github.com/uhppoted/uhppoted-httpd/httpd/controllers"
	"github.com/uhppoted/uhppoted-httpd/httpd/doors"
	"github.com/uhppoted/uhppoted-httpd/httpd/groups"
)

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	uid, role, ok := d.authorized(w, r, path)
	if !ok {
		return
	}

	if match, err := regexp.MatchString(`/system`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, "system", d.grule.system)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		controllers.Post(w, r, d.timeout, auth)
		return
	}

	if match, err := regexp.MatchString(`/doors`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, "doors", d.grule.doors)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		doors.Post(w, r, d.timeout, auth)
		return
	}

	if match, err := regexp.MatchString(`/cards`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, "cards", d.grule.cards)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		cards.Post(w, r, d.timeout, auth)
		return
	}

	if match, err := regexp.MatchString(`/groups`, path); err == nil && match {
		auth, err := NewAuthorizator(uid, role, "groups", d.grule.groups)
		if err != nil {
			http.Error(w, "Error executing request", http.StatusInternalServerError)
		}

		groups.Post(w, r, d.timeout, auth)
		return
	}

	http.Error(w, "API not implemented", http.StatusNotImplemented)
}
