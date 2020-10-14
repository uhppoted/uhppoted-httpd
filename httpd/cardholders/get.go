package cardholders

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
)

func Fetch(db db.DB, w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	var response interface{}

	go func() {
		list, err := db.CardHolders()
		if err != nil {
			warn(err)
			http.Error(w, "Error retrieving card holders", http.StatusInternalServerError)
			return
		}

		response = struct {
			DB interface{} `json:"db"`
		}{
			DB: list,
		}

		response = struct {
			DB interface{} `json:"db"`
		}{
			DB: struct {
				CardHolders interface{} `json:"cardholders"`
			}{
				CardHolders: list,
			},
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
