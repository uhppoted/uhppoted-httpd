package types

import (
	"encoding/json"
	"time"
)

type Date struct {
	ID   string    `json:"ID"`
	Date time.Time `json:"Date"`
}

func (d Date) MarshalJSON() ([]byte, error) {
	m := struct {
		ID   string `json:"ID"`
		Date string `json:"Date"`
	}{
		ID:   d.ID,
		Date: d.Date.Format("2006-01-02"),
	}

	return json.Marshal(m)
}

func (d *Date) UnmarshalJSON(bytes []byte) error {
	m := struct {
		ID   string `json:"ID"`
		Date string `json:"Date"`
	}{}

	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}

	date, err := time.ParseInLocation("2006-01-02", m.Date, time.Local)
	if err != nil {
		return err
	}

	d.ID = m.ID
	d.Date = date

	return nil
}

func (d Date) Format(layout string) string {
	return d.Date.Format(layout)
}

func (d Date) String() string {
	return d.Date.Format("2006-01-02")
}
