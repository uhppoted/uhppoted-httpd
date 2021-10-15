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
	OID        catalog.OID `json:"OID"`
	Timestamp  time.Time   `json:"timestamp"`
	UID        string      `json:"uid"`
	Module     interface{} `json:"module"`
	ModuleID   interface{} `json:"module-id"`
	ModuleName interface{} `json:"module-name"`
}

const LogTimestamp = catalog.LogTimestamp
const LogUID = catalog.LogUID
const LogModule = catalog.LogModule
const LogModuleID = catalog.LogModuleID
const LogModuleName = catalog.LogModuleName

const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID

func NewLogEntry(oid catalog.OID, timestamp time.Time, entry audit.LogEntry, lookup func(catalog.OID) interface{}) LogEntry {
	module := "controller"
	id := lookup(catalog.OID(entry.Module).Append(ControllerDeviceID))
	name := lookup(catalog.OID(entry.Module).Append(ControllerName))

	return LogEntry{
		OID:        oid,
		Timestamp:  timestamp,
		UID:        entry.UID,
		Module:     module,
		ModuleID:   id,
		ModuleName: name,
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
		catalog.NewObject2(l.OID, LogModule, l.Module),
		catalog.NewObject2(l.OID, LogModuleID, l.ModuleID),
		catalog.NewObject2(l.OID, LogModuleName, l.ModuleName),
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
