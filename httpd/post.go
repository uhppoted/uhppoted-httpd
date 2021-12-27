package httpd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/users"
	"github.com/uhppoted/uhppoted-httpd/types"
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

	case
		"/interfaces",
		"/controllers",
		"/doors",
		"/cards",
		"/groups":
		if handler := d.vtable(path); handler == nil || handler.post == nil {
			warn(fmt.Errorf("No vtable entry for %v", path))
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else if auth, err := NewAuthorizator(uid, role, handler.tag, handler.rules); err != nil {
			warn(err)
			http.Error(w, "internal system error", http.StatusInternalServerError)
		} else {
			d.dispatch(w, r, func(m map[string]interface{}) (interface{}, error) {
				return handler.post(m, auth)
			})
		}
		return

	default:
		http.Error(w, "API not implemented", http.StatusNotImplemented)
		return
	}
}

func (d *dispatcher) dispatch(w http.ResponseWriter, r *http.Request, f func(map[string]interface{}) (interface{}, error)) {
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)

	defer cancel()

	go func() {
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

		if contentType != "application/json" {
			ch <- &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid request"),
				Detail: fmt.Errorf("Invalid request content-type (%v)", contentType),
			}
			return
		}

		blob, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ch <- &types.HttpdError{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("Invalid reading request"),
				Detail: err,
			}
			return
		}

		body := map[string]interface{}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			ch <- &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid request body"),
				Detail: fmt.Errorf("Error unmarshalling request (%s): %w", string(blob), err),
			}
			return
		}

		response, err := f(body)
		if err != nil {
			ch <- err
			return
		}

		b, err := json.Marshal(response)
		if err != nil {
			ch <- &types.HttpdError{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("Internal error generating response"),
				Detail: fmt.Errorf("Error marshalling response: %w", err),
			}
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

		ch <- nil
	}()

	select {
	case <-ctx.Done():
		warn(ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return

	case err := <-ch:
		if err != nil {
			warn(err)

			switch e := err.(type) {
			case *types.HttpdError:
				http.Error(w, e.Error(), e.Status)

			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}

			return
		}
	}
}
