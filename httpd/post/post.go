package post

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/log"
)

func parseHeader(r *http.Request) (contentType string, acceptsGzip bool) {
	contentType = ""
	acceptsGzip = false

	for k, h := range r.Header {
		header := strings.TrimSpace(strings.ToLower(k))

		switch header {
		case "content-type":
			for _, v := range h {
				contentType = strings.TrimSpace(strings.ToLower(v))
			}

		case "accept-encoding":
			for _, v := range h {
				if strings.Contains(strings.TrimSpace(strings.ToLower(v)), "gzip") {
					acceptsGzip = true
				}
			}
		}
	}

	return
}

func parseRequest(r *http.Request) (map[string]any, error) {
	contentType, _ := parseHeader(r)
	body := map[string]any{}

	switch contentType {
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			return nil, err
		} else {
			for k, v := range r.Form {
				body[k] = v
			}
		}

	case "application/json":
		if blob, err := ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		} else if err := json.Unmarshal(blob, &body); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Invalid request content-type (%v)", contentType)
	}

	return body, nil
}

func get(r *http.Request, vars ...string) (map[string]string, error) {
	if body, err := parseRequest(r); err != nil {
		return nil, err
	} else {
		m := map[string]string{}
		for _, k := range vars {
			if v, ok := body[k]; !ok {
				return nil, fmt.Errorf("missing '%v'", k)
			} else if u, ok := v.([]string); ok && len(u) > 0 {
				m[k] = u[0]
			} else if u, ok := v.(string); ok {
				m[k] = u
			}
		}

		return m, nil
	}
}

func debugf(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-8v %v", subsystem, format)

	log.Debugf(f, args...)
}

func infof(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-8v %v", subsystem, format)

	log.Infof(f, args...)
}

func warnf(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-8v %v", subsystem, format)

	log.Warnf(f, args...)
}
