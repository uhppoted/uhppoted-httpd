package system

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	sys "github.com/uhppoted/uhppoted-httpd/sys"
)

func Fetch(w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	acceptsGzip := false

	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "accept-encoding" {
			for _, v := range h {
				if strings.Contains(strings.TrimSpace(strings.ToLower(v)), "gzip") {
					acceptsGzip = true
				}
			}
		}
	}

	var response interface{}

	go func() {
		object := sys.System()

		response = struct {
			System interface{} `json:"system"`
		}{
			System: object,
		}

		//     response = struct {
		//         DB interface{} `json:"db"`
		//     }{
		//         DB: struct {
		//             CardHolders interface{} `json:"cardholders"`
		//         }{
		//             CardHolders: list,
		//         },
		//     }

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

	if acceptsGzip && len(b) > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gz.Write(b)
		gz.Close()
	} else {
		w.Write(b)
	}
}
