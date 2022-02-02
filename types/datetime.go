package types

import (
	"encoding/json"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type DateTime core.DateTime

func DateTimeNow() DateTime {
	return DateTime(time.Now())
}

func DateTimePtrNow() *DateTime {
	now := DateTimeNow()

	return &now
}

func (d *DateTime) Before(t time.Time) bool {
	return (*time.Time)(d).Before(t)
}

func (d DateTime) Add(dt time.Duration) DateTime {
	return DateTime(time.Time(d).Add(dt))
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format("2006-01-02 15:04:05 MST"))
}

func (d *DateTime) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	datetime, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		datetime, err = time.ParseInLocation("2006-01-02 15:04:05 MST", s, time.Local)
		if err != nil {
			return err
		}
	}

	*d = DateTime(datetime)

	return nil
}

func (d DateTime) String() string {
	return time.Time(d).Format("2006-01-02 15:04:05")
}
