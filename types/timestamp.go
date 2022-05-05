package types

import (
	"encoding/json"
	"time"
)

type Timestamp time.Time

func TimestampNow() Timestamp {
	return Timestamp(time.Now().Truncate(1 * time.Second))
}

func (ts Timestamp) IsZero() bool {
	return time.Time(ts).IsZero()
}

// Because time.Truncate does not in any way behave like your would expect it to :-(
func (ts Timestamp) Before(t time.Time) bool {
	p := time.Time(ts).UnixMilli() / 1000
	q := t.UnixMilli() / 1000

	return p < q
}

func (ts Timestamp) Add(dt time.Duration) Timestamp {
	return Timestamp(time.Time(ts).Add(dt).Truncate(1 * time.Second))
}

func (ts Timestamp) String() string {
	if ts.IsZero() {
		return ""
	}

	return time.Time(ts).Format("2006-01-02 15:04:05 MST")
}

func (ts Timestamp) UTC() Timestamp {
	return Timestamp(time.Time(ts).UTC())
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	if ts.IsZero() {
		return json.Marshal("")
	}

	return json.Marshal(time.Time(ts).Format("2006-01-02 15:04:05 MST"))
}

func (ts *Timestamp) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	if s == "" {
		*ts = Timestamp{}
		return nil
	}

	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		timestamp, err = time.ParseInLocation("2006-01-02 15:04:05 MST", s, time.Local)
		if err != nil {
			return err
		}
	}

	*ts = Timestamp(timestamp.Truncate(1 * time.Second))

	return nil
}
