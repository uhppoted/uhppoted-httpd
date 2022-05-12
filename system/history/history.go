package history

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type History struct {
	history []Entry
}

// NTS: external to History struct because there's only ever really one of them
//      and you can't copy safely/pass as an argument with an embedded mutex
//      except by address.
var guard sync.RWMutex

func NewHistory(entries ...Entry) History {
	history := History{
		history: make([]Entry, len(entries)),
	}

	history.set(entries...)

	return history
}

// func (h *History) UseLogs(logs logs.Logs, save func()) {
// 	history := []Entry{}
// 	for _, v := range logs.List() {
// 		history = append(history, Entry{
// 			Timestamp: v.Timestamp,
// 			Item:      v.Item,
// 			ItemID:    v.ItemID,
// 			Field:     v.Field,
// 			Before:    v.Before,
// 			After:     v.After,
// 		})
// 	}
//
// 	h.set(history...)
//
// 	if save != nil {
// 		save()
// 	}
// }

func (h History) LookupController(timestamp time.Time, deviceID uint32) string {
	guard.RLock()
	defer guard.RUnlock()

	name := ""

	if deviceID != 0 {
		if oid := catalog.FindController(deviceID); oid != "" {
			if v := catalog.GetV(oid, schema.ControllerName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := h.query("controller", fmt.Sprintf("%v", deviceID), "name")

		for _, v := range edits {
			if v.Timestamp.Before(timestamp) {
				break
			}

			name = v.Before
		}
	}

	return name
}

func (h History) LookupCard(timestamp time.Time, card uint32) string {
	guard.RLock()
	defer guard.RUnlock()

	name := ""

	if card != 0 {
		if oid, ok := catalog.Find(schema.CardsOID, schema.CardNumber, card); ok && oid != "" {
			oid = oid.Trim(schema.CardNumber)
			if v := catalog.GetV(oid, schema.CardName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := h.query("card", fmt.Sprintf("%v", card), "name")

		for _, v := range edits {
			if v.Timestamp.Before(timestamp) {
				break
			}

			name = v.Before
		}
	}

	return name
}

func (h History) LookupDoor(timestamp time.Time, deviceID uint32, door uint8) string {
	guard.RLock()
	defer guard.RUnlock()

	name := ""

	if deviceID != 0 && door >= 1 && door <= 4 {
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

		edits := h.query("door", fmt.Sprintf("%v:%v", deviceID, door), "name")

		for _, v := range edits {
			if v.Timestamp.Before(timestamp) {
				break
			}

			name = v.Before
		}
	}

	return name
}

func (h *History) Load(blob json.RawMessage) error {
	guard.Lock()
	defer guard.Unlock()

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
			if e, ok := list[key]; ok && (e.Before != record.Before || e.After != record.After) {
				return fmt.Errorf("duplicate record (%#v and %#v)", record, e)
			}

			list[key] = *record
		}
	}

	history := []Entry{}
	for _, e := range list {
		history = append(history, e)
	}

	h.set(history...)

	return nil
}

func (h History) Save() (json.RawMessage, error) {
	guard.RLock()
	defer guard.RUnlock()

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
	guard.RLock()
	defer guard.RUnlock()

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

func (h *History) Received(records ...audit.AuditRecord) {
	guard.Lock()
	defer guard.Unlock()

	history := h.history

	for _, record := range records {
		unknown := time.Time{}
		timestamp := record.Timestamp
		if record.Timestamp == unknown {
			timestamp = time.Now()
		}

		history = append(history, Entry{
			Timestamp: timestamp,
			Item:      record.Component,
			ItemID:    record.Details.ID,
			Field:     record.Details.Field,
			Before:    record.Details.Before,
			After:     record.Details.After,
		})
	}

	h.set(history...)
}

func (h *History) set(list ...Entry) {
	sort.SliceStable(list, func(i, j int) bool {
		p := list[i].Timestamp
		q := list[j].Timestamp

		return q.Before(p)
	})

	h.history = list
}

func (h History) query(item, id, field string) []Entry {
	records := []Entry{}

	for _, v := range h.history {
		if v.Item == item && v.ItemID == id && v.Field == field {
			records = append(records, v)
		}
	}

	return records
}
