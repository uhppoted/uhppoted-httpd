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
		// Timestamp time.Time `json:"timestamp"`
		// UID       string
		// OID       schema.OID `json:"OID"`
		// Item      string     `json:"item"`
		// ItemID    string     `json:"id"`
		// ItemName  string     `json:"name"`
		// Field     string     `json:"field"`
		// Details   string     `json:"details"`
		// Before    string     `json:"before,omitempty"`
		// After     string     `json:"after,omitempty"`
	}{
		// Timestamp: l.Timestamp,
		// UID:       l.UID,
		// OID:       l.OID,
		// Item:      l.Item,
		// ItemID:    l.ItemID,
		// ItemName:  l.ItemName,
		// Field:     l.Field,
		// Details:   l.Details,
		// Before:    l.Before,
		// After:     l.After,
	}

	return json.Marshal(record)
}

func (e *Entry) deserialize(bytes []byte) error {
	// record := struct {
	//     Timestamp time.Time `json:"timestamp"`
	//     UID       string
	//     OID       schema.OID `json:"OID"`
	//     Item      string     `json:"item"`
	//     ItemID    string     `json:"id"`
	//     ItemName  string     `json:"name"`
	//     Field     string     `json:"field"`
	//     Details   string     `json:"details"`
	//     Before    string     `json:"before"`
	//     After     string     `json:"after"`
	// }{
	//     Timestamp: l.Timestamp,
	//     UID:       l.UID,
	//     OID:       l.OID,
	//     Item:      l.Item,
	//     ItemID:    l.ItemID,
	//     ItemName:  l.ItemName,
	//     Field:     l.Field,
	//     Details:   l.Details,
	//     Before:    l.Before,
	//     After:     l.After,
	// }

	// if err := json.Unmarshal(bytes, &record); err != nil {
	//     return err
	// }

	// l.Timestamp = record.Timestamp
	// l.UID = record.UID
	// l.OID = record.OID
	// l.Item = record.Item
	// l.ItemID = record.ItemID
	// l.ItemName = record.ItemName
	// l.Field = record.Field
	// l.Details = record.Details
	// l.Before = record.Before
	// l.After = record.After

	return nil
}
