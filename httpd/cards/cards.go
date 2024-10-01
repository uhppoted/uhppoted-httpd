package cards

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system"
)

func Get(uid, role string, rq *http.Request) any {
	start := 0
	count := math.MaxInt32

	if get := rq.FormValue("range"); get != "" {
		re := regexp.MustCompile(`([0-9]+)(?:,(\*|[0-9]+|\+[0-9]+))?`)

		if match := re.FindStringSubmatch(get); len(match) > 1 {
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

	cards := system.Cards(uid, role, start, count)

	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: cards,
	}
}

func Post(uid, role string, body map[string]interface{}) (interface{}, error) {
	updated, err := system.UpdateCards(uid, role, body)
	if err != nil {
		return nil, err
	}

	return struct {
		Cards interface{} `json:"cards"`
	}{
		Cards: updated,
	}, nil
}
