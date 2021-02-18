package system

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type datetime struct {
	DateTime *types.DateTime
	TimeZone *time.Location
	Status   status
}

type ip struct {
	IP     *address
	Status status
}

type records uint32

func (r *records) String() string {
	if r != nil {
		return fmt.Sprintf("%v", uint32(*r))
	}

	return ""
}

func timezone(s string) (*time.Location, error) {
	t, err := time.Parse("2006-01-02 15:04:05 MST", s)
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

	return nil, err
}
