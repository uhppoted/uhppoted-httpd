package history

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/logs"
)

type History struct {
	history []Entry
}

func NewHistory(entries ...Entry) History {
	history := History{
		history: make([]Entry, len(entries)),
	}

	copy(history.history, entries)

	return history
}

func (h *History) UseLogs(logs logs.Logs, save func()) {
	history := []Entry{}
	for _, v := range logs.List() {
		history = append(history, Entry{
			Timestamp: v.Timestamp,
			Item:      v.Item,
			ItemID:    v.ItemID,
			Field:     v.Field,
			Value:     v.After,
		})
	}

	h.history = history

	if save != nil {
		save()
	}
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

func (h History) LookupCard(timestamp time.Time, card uint32) string {
	if card != 0 {
		edits := h.query("card", fmt.Sprintf("%v", card), "name")

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

		if oid, ok := catalog.Find(schema.CardsOID, schema.CardNumber, card); ok && oid != "" {
			oid = oid.Trim(schema.CardNumber)
			if v := catalog.GetV(oid, schema.CardName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		return name
	}

	return ""
}

func (h History) LookupDoor(timestamp time.Time, deviceID uint32, door uint8) string {
	if deviceID != 0 && door >= 1 && door <= 4 {
		edits := h.query("door", fmt.Sprintf("%v:%v", deviceID, door), "name")

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

		if controller := catalog.FindController(deviceID); controller != "" {
			var u interface{}

			switch door {
			case 1:
				u = catalog.GetV(controller, schema.ControllerDoor1)
			case 2:
				u = catalog.GetV(controller, schema.ControllerDoor2)
			case 3:
				u = catalog.GetV(controller, schema.ControllerDoor3)

			case 4:
				u = catalog.GetV(controller, schema.ControllerDoor4)
			}

			if u != nil {
				if oid, ok := u.(schema.OID); ok {
					if v := catalog.GetV(oid, schema.DoorName); v != nil {
						name = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		return name
	}

	return ""
}

func (h *History) Load(blob json.RawMessage) error {
	f := func(bytes json.RawMessage) (*Entry, string) {
		var e Entry
		if err := e.deserialize(bytes); err != nil {
			return nil, ""
		}

		return &e, fmt.Sprintf("%v:%v", e.Timestamp, e.ItemID)
	}

	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	list := map[string]Entry{}
	for _, v := range rs {
		if record, key := f(v); record != nil {
			if e, ok := list[key]; ok && e.Value != record.Value {
				return fmt.Errorf("duplicate record (%#v and %#v)", record, e)
			}

			list[key] = *record
		}
	}

	history := []Entry{}
	for _, e := range list {
		history = append(history, e)
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

func (h History) query(item, id, field string) []Entry {
	records := []Entry{}

	for _, v := range h.history {
		if v.Item == item && v.ItemID == id && v.Field == field {
			records = append(records, v)
		}
	}

	return records
}
