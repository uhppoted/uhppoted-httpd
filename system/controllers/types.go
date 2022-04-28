package controllers

import (
	"time"
)

// // Ref. https://github.com/golang/go/issues/12388
// func timezone(s string) *time.Location {
// 	utc, _ := time.LoadLocation("UTC")
//
// 	if strings.TrimSpace(s) == "" {
// 		return time.Local
// 	}
//
// 	t, err := time.ParseInLocation("2006-01-02 15:04:05 MST", s, utc)
// 	if err == nil {
// 		return t.Location()
// 	}
//
// 	t, err = time.Parse("2006-01-02 15:04:05 -0700", s)
// 	if err == nil {
// 		_, offset := t.Zone()
// 		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset)
// 	}
//
// 	t, err = time.Parse("2006-01-02 15:04:05 Z07:00", s)
// 	if err == nil {
// 		_, offset := t.Zone()
// 		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset)
// 	}
//
// 	re := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} UTC([+-][0-9]+)")
// 	if match := re.FindStringSubmatch(s); match != nil {
// 		if offset, err := strconv.Atoi(match[1]); err == nil {
// 			return time.FixedZone(fmt.Sprintf("UTC%+d", offset), offset*3600)
// 		}
// 	}
//
// 	t, err = time.Parse("2006-01-02 15:04:05", s)
// 	if err == nil {
// 		return time.Local
// 	}
//
// 	if tz, err := time.LoadLocation(s); err == nil {
// 		return tz
// 	}
//
// 	if time.Now().Format("MST") == s {
// 		return time.Local
// 	}
//
// 	re = regexp.MustCompile("UTC([+-][0-9]{1,2})")
// 	if match := re.FindStringSubmatch(s); match != nil {
// 		if offset, err := strconv.Atoi(match[1]); err == nil {
// 			if offset != 0 {
// 				return time.FixedZone(fmt.Sprintf("UTC%+d", offset), offset*3600)
// 			}
//
// 			if tz, err := time.LoadLocation("UTC"); err == nil {
// 				return tz
// 			}
// 		}
// 	}
//
// 	// ... invalid timezone - just default to Local
// 	return time.Local
// }

func timezone(tz string) *time.Location {
	if location, err := time.LoadLocation(tz); err == nil {
		_, offset := time.Now().In(location).Zone()

		return time.FixedZone(location.String(), offset)
	}

	// ... default to Local
	return time.Local
}
