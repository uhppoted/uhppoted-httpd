package logs

import (
	"encoding/json"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LogEntry struct {
	ctypes.CatalogLogEntry
	OID       schema.OID `json:"OID"`
	Timestamp time.Time  `json:"timestamp"`
	UID       string     `json:"uid"`
	Item      string     `json:"item"`
	ItemID    string     `json:"item-id"`
	ItemName  string     `json:"item-name"`
	Field     string     `json:"field"`
	Details   string     `json:"details"`
	Before    string     `json:"before,omitempty"`
	After     string     `json:"after,omitempty"`
}

func NewLogEntry(oid schema.OID, timestamp time.Time, record audit.AuditRecord) LogEntry {
	return LogEntry{
		OID:       oid,
		Timestamp: timestamp,
		UID:       record.UID,
		Item:      record.Component,
		ItemID:    record.Details.ID,
		ItemName:  record.Details.Name,
		Field:     record.Details.Field,
		Details:   record.Details.Description,
		Before:    record.Details.Before,
		After:     record.Details.After,
	}
}

func (l LogEntry) IsValid() bool {
	return true
}

func (l LogEntry) IsDeleted() bool {
	return false
}

func (l *LogEntry) AsObjects(a auth.OpAuth) []schema.Object {
	type E = struct {
		field schema.Suffix
		value interface{}
	}

	list := []E{}

	list = append(list, E{LogTimestamp, l.Timestamp.Format(time.RFC3339)})
	list = append(list, E{LogUID, l.UID})
	list = append(list, E{LogItem, l.Item})
	list = append(list, E{LogItemID, l.ItemID})
	list = append(list, E{LogItemName, l.ItemName})
	list = append(list, E{LogField, l.Field})
	list = append(list, E{LogDetails, l.Details})

	f := func(l *LogEntry, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(l, field, value, auth.Logs); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if f(l, "OID", l.OID) {
		catalog.Join(&objects, catalog.NewObject(l.OID, types.StatusOk))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(l, field, v.value) {
			catalog.Join(&objects, catalog.NewObject2(l.OID, v.field, v.value))
		}
	}

	return objects
}

func (l *LogEntry) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Timestamp string
	}{}

	if l != nil {
		entity.Timestamp = l.Timestamp.Format("2006-01-02 15:04:05 MST")
	}

	return "log", &entity
}

func (l *LogEntry) set(auth auth.OpAuth, oid schema.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}

func (l LogEntry) serialize() ([]byte, error) {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		UID       string
		OID       schema.OID `json:"OID"`
		Item      string     `json:"item"`
		ItemID    string     `json:"id"`
		ItemName  string     `json:"name"`
		Field     string     `json:"field"`
		Details   string     `json:"details"`
		Before    string     `json:"before,omitempty"`
		After     string     `json:"after,omitempty"`
	}{
		Timestamp: l.Timestamp,
		UID:       l.UID,
		OID:       l.OID,
		Item:      l.Item,
		ItemID:    l.ItemID,
		ItemName:  l.ItemName,
		Field:     l.Field,
		Details:   l.Details,
		Before:    l.Before,
		After:     l.After,
	}

	return json.Marshal(record)
}

func (l *LogEntry) deserialize(bytes []byte) error {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		UID       string
		OID       schema.OID `json:"OID"`
		Item      string     `json:"item"`
		ItemID    string     `json:"id"`
		ItemName  string     `json:"name"`
		Field     string     `json:"field"`
		Details   string     `json:"details"`
		Before    string     `json:"before"`
		After     string     `json:"after"`
	}{
		Timestamp: l.Timestamp,
		UID:       l.UID,
		OID:       l.OID,
		Item:      l.Item,
		ItemID:    l.ItemID,
		ItemName:  l.ItemName,
		Field:     l.Field,
		Details:   l.Details,
		Before:    l.Before,
		After:     l.After,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	l.Timestamp = record.Timestamp
	l.UID = record.UID
	l.OID = record.OID
	l.Item = record.Item
	l.ItemID = record.ItemID
	l.ItemName = record.ItemName
	l.Field = record.Field
	l.Details = record.Details
	l.Before = record.Before
	l.After = record.After

	return nil
}

func (l LogEntry) clone() LogEntry {
	log := LogEntry{
		OID:       l.OID,
		Timestamp: l.Timestamp,
	}

	return log
}
