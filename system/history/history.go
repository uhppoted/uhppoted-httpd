package history

import (
	// "crypto/sha1"
	"encoding/json"
	"fmt"
	// "os"
	"sort"
	// "sync"
	"time"

	// "github.com/uhppoted/uhppoted-httpd/audit"
	// "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type History struct {
	history []Entry
	// logs map[key]LogEntry
}

type key [20]byte

// var guard sync.RWMutex

// func newKey(timestamp time.Time, uid, item, id, name, field, details string) key {
//     r := struct {
//         Timestamp time.Time `json:"timestamp"`
//         UID       string    `json:"uid"`
//         Item      string    `json:"item"`
//         ID        string    `json:"id"`
//         Name      string    `json:"name"`
//         Field     string    `json:"field"`
//         Details   string    `json:"details"`
//     }{
//         Timestamp: timestamp,
//         UID:       uid,
//         Item:      item,
//         ID:        id,
//         Name:      name,
//         Field:     field,
//         Details:   details,
//     }

//     b, _ := json.Marshal(r)
//     hash := sha1.Sum(b)

//     return key(hash)
// }

func NewHistory(entries ...Entry) History {
	history := History{
		history: make([]Entry, len(entries)),
	}

	copy(history.history, entries)

	return history
}

func (h History) LookupController(timestamp time.Time, deviceID uint32) string {
	if deviceID != 0 {
		edits := h.query("controller", fmt.Sprintf("%v", deviceID), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return p.Before(q)
		})

		name := ""
		for _, v := range edits {
			if v.Timestamp.After(timestamp) {
				return name
			}

			name = v.Value
		}

		if oid := catalog.FindController(deviceID); oid != "" {
			if v := catalog.GetV(oid, schema.ControllerName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		return name
	}

	return ""
}

func (h *History) Load(blob json.RawMessage) error {
	f := func(bytes json.RawMessage) (*Entry, key) {
		var e Entry
		if err := e.deserialize(bytes); err != nil {
			return nil, key{}
		}

		return &e, key{}
	}

	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	history := []Entry{}
	for _, v := range rs {
		if record, _ := f(v); record != nil {
			// if e, ok := logs[k]; ok {
			//     return fmt.Errorf("duplicate record (%#v and %#v)", record, x)
			// } else {
			//     logs[k] = *record
			// }
			history = append(history, *record)
		}
	}

	h.history = history

	return nil
}

func (h History) Save() (json.RawMessage, error) {
	if err := h.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}

	for _, e := range h.history {
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (h History) Print() {
	serializable := []json.RawMessage{}
	for _, e := range h.history {
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- HISTORY\n%s\n", string(b))
	}
}

func (h History) Validate() error {
	return nil
}

// func (ll *Logs) Received(records ...audit.AuditRecord) {
//     for _, record := range records {
//         unknown := time.Time{}
//         timestamp := record.Timestamp
//         if record.Timestamp == unknown {
//             timestamp = time.Now()
//         }
//
//         guard.Lock()
//         defer guard.Unlock()
//
//         k := newKey(record.Timestamp,
//             record.UID,
//             record.Component,
//             record.Details.ID,
//             record.Details.Name,
//             record.Details.Field,
//             record.Details.Description)
//
//         if _, ok := ll.logs[k]; !ok {
//             oid := catalog.NewT(LogEntry{}.CatalogLogEntry)
//             ll.logs[k] = NewLogEntry(oid, timestamp, record)
//         }
//     }
// }

// func load(file string) ([]json.RawMessage, error) {
//     blob := map[string][]json.RawMessage{}

//     bytes, err := os.ReadFile(file)
//     if err != nil {
//         return nil, err
//     }

//     if err = json.Unmarshal(bytes, &blob); err != nil {
//         return nil, err
//     }

//     return blob["logs"], nil
// }

func (h History) query(item, id, field string) []Entry {
	records := []Entry{}

	for _, v := range h.history {
		if v.Item == item && v.ItemID == id && v.Field == field {
			records = append(records, v)
		}
	}

	return records
}
