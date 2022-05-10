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
	Value     string
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
		Value     string    `json:"value,omitempty"`
	}{
		Timestamp: e.Timestamp,
		Item:      e.Item,
		ItemID:    e.ItemID,
		Field:     e.Field,
		Value:     e.Value,
	}

	return json.Marshal(record)
}

func (e *Entry) deserialize(bytes []byte) error {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		Item      string    `json:"item"`
		ItemID    string    `json:"id"`
		Field     string    `json:"field"`
		Value     string    `json:"value"`
	}{
		Timestamp: e.Timestamp,
		Item:      e.Item,
		ItemID:    e.ItemID,
		Field:     e.Field,
		Value:     e.Value,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	e.Timestamp = record.Timestamp
	e.Item = record.Item
	e.ItemID = record.ItemID
	e.Field = record.Field
	e.Value = record.Value

	return nil
}
