package types

import (
	"encoding/json"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type DateTime core.DateTime

func DateTimeNow() DateTime {
	return DateTime(time.Now().Truncate(1 * time.Second))
}

func DateTimePtrNow() *DateTime {
	now := DateTimeNow()

	return &now
}

// Because time.Truncate does not in any way behave like your would expect it to :-(
func (d DateTime) Before(t time.Time) bool {
	p := time.Time(d).UnixMilli() / 1000
	q := t.UnixMilli() / 1000

	return p < q
}

func (d DateTime) Add(dt time.Duration) DateTime {
	return DateTime(time.Time(d).Add(dt).Truncate(1 * time.Second))
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

	*d = DateTime(datetime.Truncate(1 * time.Second))

	return nil
}

func (d DateTime) String() string {
	return time.Time(d).Format("2006-01-02 15:04:05")
}
