package history

import (
	"encoding/json"
	"time"
)

type Entry struct {
	Timestamp time.Time
	Item      string
	ItemID    string
	Field     string
	Before    string
	After     string
}

func (e Entry) IsValid() bool {
	return true
}

func (e Entry) IsDeleted() bool {
	return false
}

func (e Entry) serialize() ([]byte, error) {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		Item      string    `json:"item"`
		ItemID    string    `json:"id"`
		Field     string    `json:"field"`
		Before    string    `json:"before,omitempty"`
		After     string    `json:"after,omitempty"`
	}{
		Timestamp: e.Timestamp,
		Item:      e.Item,
		ItemID:    e.ItemID,
		Field:     e.Field,
		Before:    e.Before,
		After:     e.After,
	}

	return json.Marshal(record)
}

func (e *Entry) deserialize(bytes []byte) error {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		Item      string    `json:"item"`
		ItemID    string    `json:"id"`
		Field     string    `json:"field"`
		Before    string    `json:"before"`
		After     string    `json:"after"`
	}{
		Timestamp: e.Timestamp,
		Item:      e.Item,
		ItemID:    e.ItemID,
		Field:     e.Field,
		Before:    e.Before,
		After:     e.After,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	e.Timestamp = record.Timestamp
	e.Item = record.Item
	e.ItemID = record.ItemID
	e.Field = record.Field
	e.Before = record.Before
	e.After = record.After

	return nil
}
