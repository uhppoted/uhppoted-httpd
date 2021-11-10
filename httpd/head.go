package httpd

import (
	"net/http"
)

func (d *dispatcher) head(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	return
}
