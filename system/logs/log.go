package logs

import (
	"encoding/json"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LogEntry struct {
	OID       catalog.OID `json:"OID"`
	Timestamp time.Time   `json:"timestamp"`
	UID       string      `json:"uid"`
	Item      string      `json:"item"`
	ItemID    string      `json:"item-id"`
	ItemName  string      `json:"item-name"`
	Field     string      `json:"field"`
	Details   string      `json:"details"`
	Before    string      `json:"before,omitempty"`
	After     string      `json:"after,omitempty"`
}

const LogTimestamp = catalog.LogTimestamp
const LogUID = catalog.LogUID
const LogItem = catalog.LogItem
const LogItemID = catalog.LogItemID
const LogItemName = catalog.LogItemName
const LogField = catalog.LogField
const LogDetails = catalog.LogDetails

const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID

func NewLogEntry(oid catalog.OID, timestamp time.Time, record audit.AuditRecord) LogEntry {
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

func (l *LogEntry) AsObjects() []interface{} {

	objects := []interface{}{
		catalog.NewObject(l.OID, types.StatusOk),
		catalog.NewObject2(l.OID, LogTimestamp, l.Timestamp.Format(time.RFC3339)),
		catalog.NewObject2(l.OID, LogUID, l.UID),
		catalog.NewObject2(l.OID, LogItem, l.Item),
		catalog.NewObject2(l.OID, LogItemID, l.ItemID),
		catalog.NewObject2(l.OID, LogItemName, l.ItemName),
		catalog.NewObject2(l.OID, LogField, l.Field),
		catalog.NewObject2(l.OID, LogDetails, l.Details),
	}

	return objects
}

func (l *LogEntry) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}

func (l LogEntry) serialize() ([]byte, error) {
	record := struct {
		Timestamp time.Time `json:"timestamp"`
		UID       string
		OID       catalog.OID `json:"OID"`
		Item      string      `json:"item"`
		ItemID    string      `json:"id"`
		ItemName  string      `json:"name"`
		Field     string      `json:"field"`
		Details   string      `json:"details"`
		Before    string      `json:"before,omitempty"`
		After     string      `json:"after,omitempty"`
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
		OID       catalog.OID `json:"OID"`
		Item      string      `json:"item"`
		ItemID    string      `json:"id"`
		ItemName  string      `json:"name"`
		Field     string      `json:"field"`
		Details   string      `json:"details"`
		Before    string      `json:"before"`
		After     string      `json:"after"`
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
