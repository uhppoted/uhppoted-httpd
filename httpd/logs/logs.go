package logs

import (
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system"
)

const GZIP_MINIMUM = 16384

func Get(uid, role string, rq *http.Request) interface{} {
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

	return struct {
		Logs interface{} `json:"logs"`
	}{
		Logs: system.Logs(uid, role, start, count),
	}
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
