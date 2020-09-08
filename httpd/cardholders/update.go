package cardholders

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
)

func Update(db db.DB, w http.ResponseWriter, r *http.Request, timeout time.Duration) {
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

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		defer cancel()

		var response interface{}

		go func() {
			updated, err := db.Update(body)
			if err != nil {
				warn(err)
				http.Error(w, "Error updating card holders", http.StatusInternalServerError)
				return
			}

			response = struct {
				DB interface{} `json:"db"`
			}{
				DB: updated,
			}

			cancel()
		}()

		<-ctx.Done()

		if err := ctx.Err(); err != context.Canceled {
			warn(err)
			http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
			return
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
