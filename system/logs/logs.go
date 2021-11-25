package logs

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Logs struct {
	Logs map[key]LogEntry `json:"logs"`
}

type key [20]byte

const LogsOID = catalog.LogsOID
const LogsFirst = catalog.LogsFirst
const LogsLast = catalog.LogsLast

var guard sync.RWMutex

func newKey(timestamp time.Time, uid, item, id, name, field, details string) key {
	r := struct {
		Timestamp time.Time `json:"timestamp"`
		UID       string    `json:"uid"`
		Item      string    `json:"item"`
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Field     string    `json:"field"`
		Details   string    `json:"details"`
	}{
		Timestamp: timestamp,
		UID:       uid,
		Item:      item,
		ID:        id,
		Name:      name,
		Field:     field,
		Details:   details,
	}

	b, _ := json.Marshal(r)
	hash := sha1.Sum(b)

	return key(hash)
}

func NewLogs() Logs {
	logs := Logs{
		Logs: map[key]LogEntry{},
	}

	return logs
}

func (ll *Logs) Load(blob json.RawMessage) error {
	f := func(bytes json.RawMessage) (*LogEntry, key) {
		var l LogEntry
		if err := l.deserialize(bytes); err != nil {
			return nil, key{}
		}

		return &l, newKey(l.Timestamp, l.UID, l.Item, l.ItemID, l.ItemName, l.Field, l.Details)
	}

	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	logs := map[key]LogEntry{}
	for _, v := range rs {
		if record, k := f(v); record != nil {
			if x, ok := logs[k]; ok {
				return fmt.Errorf("duplicate record (%#v and %#v)", record, x)
			} else {
				logs[k] = *record
			}
		}
	}

	for _, v := range logs {
		catalog.PutLogEntry(v.OID)
	}

	ll.Logs = logs

	return nil
}

func (ll Logs) Save() (json.RawMessage, error) {
	if err := validate(ll); err != nil {
		return nil, err
	}

	if err := scrub(ll); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}

	for _, l := range ll.Logs {
		if l.IsValid() && !l.IsDeleted() {
			if record, err := l.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (ll Logs) Print() {
	if b, err := json.MarshalIndent(ll.Logs, "", "  "); err == nil {
		fmt.Printf("----------------- LOGS\n%s\n", string(b))
	}
}

func (ll *Logs) Clone() *Logs {
	shadow := Logs{
		Logs: map[key]LogEntry{},
	}

	for k, l := range ll.Logs {
		shadow.Logs[k] = l.clone()
	}

	return &shadow
}

func (ll *Logs) AsObjects(start, max int) []interface{} {
	guard.RLock()
	defer guard.RUnlock()

	objects := []interface{}{}
	keys := []key{}

	for k := range ll.Logs {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		p := ll.Logs[keys[i]].Timestamp
		q := ll.Logs[keys[j]].Timestamp

		return q.Before(p)
	})

	ix := start
	count := 0
	for ix < len(keys) && count < max {
		k := keys[ix]
		if l, ok := ll.Logs[k]; ok {
			if l.IsValid() || l.IsDeleted() {
				if l := l.AsObjects(); l != nil {
					objects = append(objects, l...)
					count++
				}
			}
		}

		ix++
	}

	if len(keys) > 0 {
		first := ll.Logs[keys[0]]
		last := ll.Logs[keys[len(keys)-1]]
		objects = append(objects, catalog.NewObject2(LogsOID, LogsFirst, first.OID))
		objects = append(objects, catalog.NewObject2(LogsOID, LogsLast, last.OID))
	}

	return objects
}

func (ll *Logs) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	if ll == nil {
		return nil, nil
	}

	for k, e := range ll.Logs {
		if e.OID.Contains(oid) {
			objects, err := e.set(auth, oid, value)
			if err == nil {
				ll.Logs[k] = e
			}

			return objects, err
		}
	}

	return []interface{}{}, nil
}

func (ll *Logs) Validate() error {
	return nil
}

func (ll *Logs) Received(records ...audit.AuditRecord) {
	for _, record := range records {
		unknown := time.Time{}
		timestamp := record.Timestamp
		if record.Timestamp == unknown {
			timestamp = time.Now()
		}

		guard.Lock()
		defer guard.Unlock()

		k := newKey(record.Timestamp, record.UID, record.Component, record.Details.ID, record.Details.Name, record.Details.Field, record.Details.Description)
		if _, ok := ll.Logs[k]; !ok {
			oid := catalog.NewLogEntry()
			ll.Logs[k] = NewLogEntry(oid, timestamp, record)
		}
	}
}

func (ll *Logs) Query(item, id, field string) []LogEntry {
	records := []LogEntry{}

	for _, v := range ll.Logs {
		if v.Item == item && v.ItemID == id && v.Field == field {
			records = append(records, v)
		}
	}

	return records
}

func load(file string) ([]json.RawMessage, error) {
	blob := map[string][]json.RawMessage{}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &blob); err != nil {
		return nil, err
	}

	return blob["logs"], nil
}

func validate(ll Logs) error {
	return nil
}

func scrub(ll Logs) error {
	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
