package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Ref. https://github.com/golang/go/issues/12388
func Timezone(s string) (*time.Location, error) {
	utc, _ := time.LoadLocation("UTC")

	if strings.TrimSpace(s) == "" {
		return time.Local, nil
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05 MST", strings.ToUpper(s), utc)
	if err == nil {
		return t.Location(), nil
	}

	t, err = time.ParseInLocation("2006-01-02 15:04 MST", strings.ToUpper(s), utc)
	if err == nil {
		return t.Location(), nil
	}

	t, err = time.Parse("2006-01-02 15:04:05 -0700", s)
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	t, err = time.Parse("2006-01-02 15:04 -0700", s)
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	t, err = time.Parse("2006-01-02 15:04:05 Z07:00", strings.ToUpper(s))
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	t, err = time.Parse("2006-01-02 15:04 Z07:00", strings.ToUpper(s))
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	re := regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}(?:\:[0-9]{2})?\s+(.+)`)
	if match := re.FindStringSubmatch(s); match != nil {
		if tz, err := time.LoadLocation(match[1]); err == nil {
			_, offset := time.Now().In(tz).Zone()

			return time.FixedZone(tz.String(), offset), nil
		}
	}

	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return time.Local, nil
	}

	t, err = time.Parse("2006-01-02 15:04", s)
	if err == nil {
		return time.Local, nil
	}

	// e.g. Africa/Cairo
	if tz, err := time.LoadLocation(s); err == nil {
		_, offset := time.Now().In(tz).Zone()

		return time.FixedZone(tz.String(), offset), nil
	}

	re = regexp.MustCompile("UTC([+-][0-9]{1,2})")
	if match := re.FindStringSubmatch(strings.ToUpper(s)); match != nil {
		if offset, err := strconv.Atoi(match[1]); err == nil {
			if offset != 0 {
				return time.FixedZone(fmt.Sprintf("UTC%+d", offset), offset*3600), nil
			}

			if tz, err := time.LoadLocation("UTC"); err == nil {
				return tz, nil
			}
		}
	}

	// Workaround for 'PDT unknown timezone' (ref. https://github.com/golang/go/issues/12388)
	now := fmt.Sprintf("%v %v", time.Now().Format("2006-01-02 15:04:05"), s)
	if t, err := time.ParseInLocation("2006-01-02 15:04:05 MST", now, time.Local); err == nil {
		zone, offset := time.Now().In(t.Location()).Zone()

		return time.FixedZone(zone, offset), nil
	}

	return nil, fmt.Errorf("Invalid timezone (%v)", s)
}
