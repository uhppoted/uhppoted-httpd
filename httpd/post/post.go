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

func parseRequest(r *http.Request, contentType string) (map[string]any, error) {
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

func get(body map[string]any, key string) (string, error) {
	if v, ok := body[key]; !ok {
		return "", fmt.Errorf("missing '%v'", key)
	} else if u, ok := v.([]string); ok && len(u) > 0 {
		return u[0], nil
	} else if u, ok := v.(string); ok {
		return u, nil
	}

	return "", fmt.Errorf("invalid '%v'", key)
}

func debugf(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-3v  %v", subsystem, format)

	log.Debugf(f, args...)
}

func infof(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-3v  %v", subsystem, format)

	log.Infof(f, args...)
}

func warnf(subsystem string, format string, args ...any) {
	f := fmt.Sprintf("%-3v  %v", subsystem, format)

	log.Warnf(f, args...)
}
