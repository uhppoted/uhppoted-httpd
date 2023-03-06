package httpd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

func (d *dispatcher) exec(w http.ResponseWriter, r *http.Request, f func(map[string]interface{}) (interface{}, error)) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		defer close(ch)

		acceptsGzip := false
		contentType := ""

		for k, h := range r.Header {
			if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
				for _, v := range h {
					contentType = strings.TrimSpace(strings.ToLower(v))
				}
			}

			if strings.TrimSpace(strings.ToLower(k)) == "accept-encoding" {
				for _, v := range h {
					if strings.Contains(strings.TrimSpace(strings.ToLower(v)), "gzip") {
						acceptsGzip = true
					}
				}
			}
		}

		body := map[string]interface{}{}

		switch contentType {
		case "application/x-www-form-urlencoded":
			if err := r.ParseForm(); err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Error reading request", http.StatusInternalServerError)
				return
			}

			for k, v := range r.Form {
				body[k] = v
			}

		case "application/json":
			blob, err := io.ReadAll(r.Body)
			if err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Error reading request", http.StatusInternalServerError)
				return
			}

			if err := json.Unmarshal(blob, &body); err != nil {
				warnf("HTTPD", "%v", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, fmt.Sprintf("Invalid request content-type (%v)", contentType), http.StatusBadRequest)
			return
		}

		response, err := f(body)
		if err != nil && errors.Is(err, auth.ErrUnauthorised) {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := json.Marshal(response)
		if err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, "Internal error generating response", http.StatusInternalServerError)
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
	}()

	select {
	case <-ctx.Done():
		warnf("HTTPD", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return

	case <-ch:
	}
}

func (d *dispatcher) exec2(w http.ResponseWriter, r *http.Request, f func() (any, error)) {
	acceptsGzip := parseHeader(r)
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	go func() {
		defer close(ch)

		if response, err := f(); err != nil {
			warnf("HTTPD", "%v", err)
		} else if response == nil {
			// nothing to do
		} else if b, err := json.Marshal(response); err != nil {
			warnf("HTTPD", "%v", err)
			http.Error(w, "Internal error generating response", http.StatusInternalServerError)
		} else {
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
	}()

	select {
	case <-ctx.Done():
		warnf("HTTPD", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return

	case <-ch:
	}
}
