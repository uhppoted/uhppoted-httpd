package types

import (
	"encoding/json"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type DateTime core.DateTime

func (d *DateTime) Copy() *DateTime {
	if d == nil {
		return nil
	}

	datetime := *d

	return &datetime
}

func (d *DateTime) IsValid() bool {
	if d != nil {
		return true
	}

	return false
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02 15:04:05 MST"))
}

func (d *DateTime) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	datetime, err := time.ParseInLocation("2006-01-02 15:04:05 MST", s, time.Local)
	if err != nil {
		return err
	}

	*d = DateTime(datetime)

	return nil
}

func (d *DateTime) Format(layout string) string {
	if d != nil {
		return time.Time(*d).Format(layout)
	}

	return ""
}

func (d *DateTime) String() string {
	if d != nil {
		return time.Time(*d).Format("2006-01-02 15:04:05")
	}

	return ""
}
