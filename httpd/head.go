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
		d.authenticate(w, r)
		return
	}

	http.Error(w, "invalid URL", http.StatusNotFound)
}
