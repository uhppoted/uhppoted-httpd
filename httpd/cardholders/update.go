package cardholders

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
	"github.com/uhppoted/uhppoted-httpd/sys"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Update(db db.DB, w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	go func() {
		match := regexp.MustCompile(`/cardholders/(\S+)`).FindStringSubmatch(r.URL.Path)
		if match == nil || len(match) < 2 {
			ch <- &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid request"),
				Detail: fmt.Errorf("Missing request record ID (%v)", r.URL.Path),
			}
			return
		}

		id := match[1]

		var contentType string

		for k, h := range r.Header {
			if strings.TrimSpace(strings.ToLower(k)) == "content-type" {
				for _, v := range h {
					contentType = strings.TrimSpace(strings.ToLower(v))
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

		updated, err := db.Post(id, body)
		if err != nil {
			ch <- err
			return
		}

		response := struct {
			DB interface{} `json:"db"`
		}{
			DB: updated,
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
		w.Write(b)
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

	acl, err := db.ACL()
	if err != nil {
		warn(err)
		return
	}

	system.Update(acl)
}
