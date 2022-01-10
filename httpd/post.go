package httpd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/users"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request) {
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	// ... allow unauthenticated access to /authenticate and /logout
	if path == "/authenticate" {
		d.authenticate(w, r)
		return
	}

	if path == "/logout" {
		d.logout(w, r)
		return
	}

	// ... require auth for everything else
	uid, role, authenticated := d.authenticated(r)
	if !authenticated {
		d.unauthenticated(r, w)
		return
	}

	if ok := d.authorised(uid, role, path); !ok {
		d.unauthorised(r, w)
		return
	}

	switch path {
	case "/password":
		d.dispatch(w, r, func(m map[string]interface{}) (interface{}, error) {
			return users.Post(m, r, d.auth)
		})
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
		} else {
			auth := auth.NewAuthorizator(uid, role)
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
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

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

		body := map[string]interface{}{}

		switch contentType {
		case "application/x-www-form-urlencoded":
			if err := r.ParseForm(); err != nil {
				ch <- &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid reading request"),
					Detail: err,
				}
				return
			}

			for k, v := range r.Form {
				body[k] = v
			}

		case "application/json":
			blob, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ch <- &types.HttpdError{
					Status: http.StatusInternalServerError,
					Err:    fmt.Errorf("Invalid reading request"),
					Detail: err,
				}
				return
			}

			if err := json.Unmarshal(blob, &body); err != nil {
				ch <- &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid request body"),
					Detail: fmt.Errorf("Error unmarshalling request (%s): %w", string(blob), err),
				}
				return
			}

		default:
			ch <- &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid request"),
				Detail: fmt.Errorf("Invalid request content-type (%v)", contentType),
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
