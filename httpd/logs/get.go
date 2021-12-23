package logs

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system"
)

func Fetch(w http.ResponseWriter, rq *http.Request, timeout time.Duration) {
	start := 0
	count := math.MaxInt32

	if get := rq.FormValue("range"); get != "" {
		re := regexp.MustCompile(`([0-9]+)(?:,(\*|[0-9]+|\+[0-9]+))?`)

		if match := re.FindStringSubmatch(get); match != nil && len(match) > 1 {
			if v, err := strconv.ParseUint(match[1], 10, 32); err == nil {
				start = int(v)
			}

			if len(match) > 2 {
				switch {
				case strings.TrimSpace(match[2]) == "*":
					count = math.MaxInt32

				case strings.HasPrefix(strings.TrimSpace(match[2]), "+"):
					if v, err := strconv.ParseUint(match[2][1:], 10, 32); err == nil {
						count = int(v)
					}

				default:
					if v, err := strconv.ParseUint(match[2], 10, 32); err == nil {
						count = int(v) - start
					}
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	acceptsGzip := false

	for k, h := range rq.Header {
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
		response = struct {
			Logs interface{} `json:"logs"`
		}{
			Logs: system.Logs(start, count),
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

	if acceptsGzip && len(b) > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gz.Write(b)
		gz.Close()
	} else {
		w.Write(b)
	}
}
