package get

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
	"github.com/uhppoted/uhppoted-httpd/log"
)

const GZIP_MINIMUM = 16384

func parseHeader(r *http.Request) bool {
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

	return acceptsGzip
}

func parseSettings(r *http.Request) string {
	theme := "default"

	if cookie, err := r.Cookie(cookies.SettingsCookie); err == nil {
		re := regexp.MustCompile("(.*?):(.+)")
		tokens := strings.Split(cookie.Value, ",")
		for _, token := range tokens {
			match := re.FindStringSubmatch(token)
			if len(match) > 2 {
				if strings.TrimSpace(match[1]) == "theme" {
					theme = strings.TrimSpace(match[2])
				}
			}
		}
	}

	return theme
}

func debugf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Debugf("%v", args...)
	} else {
		log.Debugf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func infof(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Infof("%v", args...)
	} else {
		log.Infof(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func warnf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Warnf("%v", args...)
	} else {
		log.Warnf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}
