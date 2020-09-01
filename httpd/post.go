package httpd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
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

	switch path {
	case "/logout":
		d.logout(w, r)

	case "/update":
		d.update(w, r)

	default:
		http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
	}
}

func (d *dispatcher) update(w http.ResponseWriter, r *http.Request) {
	var contentType string

	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
			for _, v := range h {
				contentType = strings.TrimSpace(strings.ToLower(v))
			}
		}
	}

	if contentType == "application/json" {
		blob, err := ioutil.ReadAll(r.Body)
		if err != nil {
			warn(err)
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		body := map[string]interface{}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			warn(err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updated, err := d.db.Update(body)
		if err != nil {
			warn(err)
			http.Error(w, "Error updating data", http.StatusInternalServerError)
			return
		}

		response := struct {
			DB interface{} `json:"db"`
		}{
			DB: updated,
		}

		b, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Error generating response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}
