package types

import (
	"encoding/json"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type Date core.Date

func ParseDate(s string) (*Date, error) {
	if date, err := time.ParseInLocation("2006-01-02", s, time.Local); err != nil {
		return nil, err
	} else {
		return (*Date)(&date), nil
	}
}

func (d *Date) Copy() *Date {
	if d == nil {
		return nil
	}

	date := *d

	return &date
}

func (d *Date) IsValid() bool {
	if d == nil || time.Time(*d).IsZero() {
		return false
	}

	return true
}

func (d Date) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return json.Marshal("")
	} else {
		return json.Marshal(d.Format("2006-01-02"))
	}
}

func (d *Date) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	date, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return err
	}

	*d = Date(date)

	return nil
}

func (d *Date) Format(layout string) string {
	if d != nil {
		return time.Time(*d).Format(layout)
	}

	return ""
}

func (d *Date) String() string {
	if d == nil {
		return ""
	}

	if time.Time(*d).IsZero() {
		return ""
	}

	return time.Time(*d).Format("2006-01-02")
}
