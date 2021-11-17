package users

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Password(w http.ResponseWriter, r *http.Request, timeout time.Duration, auth auth.OpAuth) {
	var uid string
	var old string
	var pwd string
	var pwd2 string
	var contentType string
	var acceptsGzip bool

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

	switch contentType {
	case "application/x-www-form-urlencoded":
		uid = r.FormValue("uid")
		old = r.FormValue("old")
		pwd = r.FormValue("pwd")
		pwd2 = r.FormValue("pwd2")

	case "application/json":
		blob, err := io.ReadAll(r.Body)
		if err != nil {
			warn(err)
			http.Error(w, "Error reading request", http.StatusInternalServerError)
			return
		}

		body := struct {
			UID  string `json:"uid"`
			Old  string `json:"old"`
			Pwd  string `json:"pwd"`
			Pwd2 string `json:"pwd2"`
		}{}

		if err := json.Unmarshal(blob, &body); err != nil {
			warn(err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		uid = body.Old
		old = body.Old
		pwd = body.Pwd
		pwd2 = body.Pwd2
	}

	// ... update users
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	go func() {
		rq := map[string]interface{}{
			"uid":  uid,
			"old":  old,
			"pwd":  pwd,
			"pwd2": pwd2,
		}

		_, err := system.UpdateUsers(rq, auth)
		if err != nil {
			ch <- err
			return
		}

		response := struct {
		}{}

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
