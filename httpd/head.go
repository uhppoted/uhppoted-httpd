package httpd

import (
	"net/http"
)

func (d *dispatcher) head(w http.ResponseWriter, r *http.Request) {
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	if path == "/authenticate" {
		d.preauthenticate(w)
		return
	}

	http.Error(w, "invalid URL", http.StatusNotFound)
}

func (d *dispatcher) preauthenticate(w http.ResponseWriter) {
	cookie, err := d.auth.Preauthenticate()
	if err != nil {
		warn(err)
		http.Error(w, "Error generating login token", http.StatusInternalServerError)
		return
	}

	if cookie != nil {
		http.SetCookie(w, cookie)
	}
}
