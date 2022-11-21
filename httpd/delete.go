package httpd

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/users"
)

func (d *dispatcher) delete(w http.ResponseWriter, r *http.Request) {
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	// ... auth
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
	case "/otp":
		users.RevokeOTP(uid, role, w, r, d.auth)

	default:
		http.Error(w, "API not implemented", http.StatusNotImplemented)
	}
}
