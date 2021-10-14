package logs

import (
	"encoding/json"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LogEntry struct {
	OID       catalog.OID `json:"OID"`
	Timestamp time.Time   `json:"timestamp"`
}

const LogTimestamp = catalog.LogTimestamp

func NewLogEntry(oid catalog.OID, timestamp time.Time) LogEntry {
	return LogEntry{
		OID:       oid,
		Timestamp: timestamp,
	}
}

func (l LogEntry) IsValid() bool {
	return true
}

func (l LogEntry) IsDeleted() bool {
	return false
}

func (l *LogEntry) AsObjects() []interface{} {
	timestamp := l.Timestamp.Format(time.RFC3339)

	objects := []interface{}{
		catalog.NewObject(l.OID, types.StatusOk),
		catalog.NewObject2(l.OID, LogTimestamp, timestamp),
	}

	return objects
}

func (l *LogEntry) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}

func (l LogEntry) serialize() ([]byte, error) {
	record := struct {
		OID       catalog.OID `json:"OID"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		OID:       l.OID,
		Timestamp: l.Timestamp,
	}

	return json.Marshal(record)
}

func (l *LogEntry) deserialize(bytes []byte) error {
	record := struct {
		OID       catalog.OID `json:"OID"`
		Timestamp time.Time   `json:"timestamp"`
	}{}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	l.OID = record.OID
	l.Timestamp = record.Timestamp

	return nil
}

func (l LogEntry) clone() LogEntry {
	log := LogEntry{
		OID:       l.OID,
		Timestamp: l.Timestamp,
	}

	return log
}
