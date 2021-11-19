package httpd

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/cards"
	"github.com/uhppoted/uhppoted-httpd/httpd/controllers"
	"github.com/uhppoted/uhppoted-httpd/httpd/doors"
	"github.com/uhppoted/uhppoted-httpd/httpd/groups"
	"github.com/uhppoted/uhppoted-httpd/httpd/users"
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

	switch path {
	case "/password":
		users.Password(w, r, d.timeout, d.auth)
		return

	case "/system":
		if auth, err := NewAuthorizator(uid, role, "system", d.grule.system); err != nil {
			warn(err)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else {
			controllers.Post(w, r, d.timeout, auth)
		}
		return

	case "/doors":
		if auth, err := NewAuthorizator(uid, role, "doors", d.grule.doors); err != nil {
			warn(err)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else {
			doors.Post(w, r, d.timeout, auth)
		}
		return

	case "/cards":
		if auth, err := NewAuthorizator(uid, role, "cards", d.grule.cards); err != nil {
			warn(err)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else {
			cards.Post(w, r, d.timeout, auth)
		}
		return

	case "/groups":
		if auth, err := NewAuthorizator(uid, role, "groups", d.grule.groups); err != nil {
			warn(err)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else {
			groups.Post(w, r, d.timeout, auth)
		}
		return

	default:
		http.Error(w, "API not implemented", http.StatusNotImplemented)
		return
	}
}
