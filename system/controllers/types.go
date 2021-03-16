package controllers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type datetime struct {
	Qwerty   int
	Expected *types.DateTime
	DateTime *types.DateTime
	Status   status
}

func (dt datetime) String() string {
	if dt.DateTime == nil || dt.Status == StatusUnknown {
		return ""
	}

	return dt.DateTime.Format("2006-01-02 15:04 MST")
}

func (dt datetime) MarshalJSON() ([]byte, error) {
	object := struct {
		DateTime string `json:"DateTime"`
		Status   string `json:"Status"`
		Expected string `json:"Expected"`
	}{
		DateTime: dt.String(),
		Status:   dt.Status.String(),
		Expected: dt.Expected.Format("2006-01-02 15:04 MST"),
	}

	return json.Marshal(object)
}

type ip struct {
	Configured *types.Address
	Address    *types.Address
	Status     status
}

type cards struct {
	Records records
	Status  status
}

func (c cards) String() string {
	if c.Status == StatusUnknown {
		return ""
	}

	return fmt.Sprintf("%v", c.Records)
}

type records uint32

func (r *records) String() string {
	if r != nil {
		return fmt.Sprintf("%v", uint32(*r))
	}

	return ""
}

// Ref. https://github.com/golang/go/issues/12388
func timezone(s string) (*time.Location, error) {
	utc, _ := time.LoadLocation("UTC")

	if strings.TrimSpace(s) == "" {
		return time.Local, nil
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05 MST", s, utc)
	if err == nil {
		return t.Location(), nil
	}

	t, err = time.Parse("2006-01-02 15:04:05 -0700", s)
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	t, err = time.Parse("2006-01-02 15:04:05 Z07:00", s)
	if err == nil {
		_, offset := t.Zone()
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	re := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} UTC([+-][0-9]+)")
	if match := re.FindStringSubmatch(s); match != nil {
		if offset, err := strconv.Atoi(match[1]); err == nil {
			return time.FixedZone(fmt.Sprintf("UTC%+d", offset), offset*3600), nil
		}
	}

	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return time.Local, nil
	}

	re = regexp.MustCompile("UTC([+-][0-9]{1,2})")
	if tz, err := time.LoadLocation(s); err == nil {
		return tz, nil
	}

	if match := re.FindStringSubmatch(s); match != nil {
		if offset, err := strconv.Atoi(match[1]); err == nil {
			if offset != 0 {
				return time.FixedZone(fmt.Sprintf("UTC%+d", offset), offset*3600), nil
			}

			if tz, err := time.LoadLocation("UTC"); err == nil {
				return tz, nil
			}
		}
	}

	return nil, fmt.Errorf("Invalid timezone (%v)", s)
}
