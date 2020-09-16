package types

import (
	"encoding/json"
	"time"
)

type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format("2006-01-02"))
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

func (d Date) Format(layout string) string {
	return time.Time(d).Format(layout)
}

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}
