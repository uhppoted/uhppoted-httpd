package logs

import (
	//	"encoding/json"
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LogEntry struct {
	OID           catalog.OID `json:"OID"`
	Timestamp     time.Time   `json:"timestamp"`
	UID           string      `json:"uid"`
	Component     interface{} `json:"component"`
	ComponentID   interface{} `json:"component-id"`
	ComponentName interface{} `json:"component-name"`
	ModuleField   interface{} `json:"module-field"`
	Details       interface{} `json:"details"`
}

const LogTimestamp = catalog.LogTimestamp
const LogUID = catalog.LogUID
const LogComponent = catalog.LogComponent
const LogComponentID = catalog.LogComponentID
const LogComponentName = catalog.LogComponentName
const LogModuleField = catalog.LogModuleField
const LogDetails = catalog.LogDetails

const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID

func NewLogEntry(oid catalog.OID, timestamp time.Time, entry audit.LogEntry) LogEntry {
	component := entry.Component
	id := entry.Info.ID()
	name := entry.Info.Name()
	field := entry.Info.Field()
	details := entry.Info.Details()

	return LogEntry{
		OID:           oid,
		Timestamp:     timestamp,
		UID:           entry.UID,
		Component:     component,
		ComponentID:   id,
		ComponentName: name,
		ModuleField:   field,
		Details:       details,
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
		catalog.NewObject2(l.OID, LogComponent, l.Component),
		catalog.NewObject2(l.OID, LogComponentID, l.ComponentID),
		catalog.NewObject2(l.OID, LogComponentName, l.ComponentName),
		catalog.NewObject2(l.OID, LogModuleField, l.ModuleField),
		catalog.NewObject2(l.OID, LogDetails, l.Details),
	}

	return objects
}

func (l *LogEntry) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}

func (l LogEntry) serialize() ([]byte, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED")
	//	record := struct {
	//		OID       catalog.OID `json:"OID"`
	//		Timestamp time.Time   `json:"timestamp"`
	//	}{
	//		OID:       l.OID,
	//		Timestamp: l.Timestamp,
	//	}
	//
	//	return json.Marshal(record)
}

func (l *LogEntry) deserialize(bytes []byte) error {
	return fmt.Errorf("NOT IMPLEMENTED")
	//	record := struct {
	//		OID       catalog.OID `json:"OID"`
	//		Timestamp time.Time   `json:"timestamp"`
	//	}{}
	//
	//	if err := json.Unmarshal(bytes, &record); err != nil {
	//		return err
	//	}
	//
	//	l.OID = record.OID
	//	l.Timestamp = record.Timestamp
	//
	//	return nil
}

func (l LogEntry) clone() LogEntry {
	log := LogEntry{
		OID:       l.OID,
		Timestamp: l.Timestamp,
	}

	return log
}
