package events

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Events struct {
	Events map[key]Event `json:"events"`

	file string `json:"-"`
}

type key struct {
	deviceID  uint32
	index     uint32
	timestamp time.Time
}

const EventsOID = catalog.EventsOID
const EventsFirst = catalog.EventsFirst
const EventsLast = catalog.EventsLast

var guard sync.RWMutex

func newKey(deviceID uint32, index uint32, timestamp time.Time) key {
	year, month, day := timestamp.Date()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()
	location := timestamp.Location()
	t := time.Date(year, month, day, hour, minute, second, 0, location)

	return key{
		deviceID:  deviceID,
		index:     index,
		timestamp: t,
	}
}

func NewEvents() Events {
	return Events{
		Events: map[key]Event{},
	}
}

func (ee *Events) Load(file string) error {
	blob := struct {
		Events []json.RawMessage `json:"events"`
	}{
		Events: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &blob); err != nil {
		return err
	}

	for _, v := range blob.Events {
		var e Event
		if err := e.deserialize(v); err == nil {
			k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
			if x, ok := ee.Events[k]; ok {
				return fmt.Errorf("%v  duplicate events (%v and %v)", k, e.OID, x.OID)
			} else {
				ee.Events[k] = e
			}
		}
	}

	for _, e := range ee.Events {
		catalog.PutEvent(e.OID)
	}

	ee.file = file

	return nil
}

func (ee Events) Save() error {
	if err := validate(ee); err != nil {
		return err
	}

	if err := scrub(ee); err != nil {
		return err
	}

	if ee.file == "" {
		return nil
	}

	serializable := struct {
		Events []json.RawMessage `json:"events"`
	}{
		Events: []json.RawMessage{},
	}

	for _, e := range ee.Events {
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable.Events = append(serializable.Events, record)
			}
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-events.*")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(ee.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), ee.file)
}

func (ee *Events) Stash() {
}

func (ee Events) Print() {
	if b, err := json.MarshalIndent(ee.Events, "", "  "); err == nil {
		fmt.Printf("----------------- EVENTS\n%s\n", string(b))
	}
}

func (ee *Events) Clone() *Events {
	shadow := Events{
		Events: map[key]Event{},
		file:   ee.file,
	}

	for k, e := range ee.Events {
		shadow.Events[k] = e.clone()
	}

	return &shadow
}

func (ee *Events) AsObjects(start, max int) []interface{} {
	guard.RLock()
	defer guard.RUnlock()

	objects := []interface{}{}
	keys := []key{}

	for k := range ee.Events {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		p := keys[i]
		q := keys[j]
		return q.timestamp.Before(p.timestamp)
	})

	ix := start
	count := 0
	for ix < len(keys) && count < max {
		k := keys[ix]
		if e, ok := ee.Events[k]; ok {
			if e.IsValid() || e.IsDeleted() {
				if l := e.AsObjects(); l != nil {
					objects = append(objects, l...)
					count++
				}
			}
		}

		ix++
	}

	if len(keys) > 0 {
		first := ee.Events[keys[0]]
		last := ee.Events[keys[len(keys)-1]]
		objects = append(objects, catalog.NewObject2(EventsOID, EventsFirst, first.OID))
		objects = append(objects, catalog.NewObject2(EventsOID, EventsLast, last.OID))

	}

	return objects
}

func (ee *Events) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	if ee == nil {
		return nil, nil
	}

	for k, e := range ee.Events {
		if e.OID.Contains(oid) {
			objects, err := e.set(auth, oid, value)
			if err == nil {
				ee.Events[k] = e
			}

			return objects, err
		}
	}

	return []interface{}{}, nil
}

func (ee *Events) Validate() error {
	return nil
}

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	guard.Lock()
	defer guard.Unlock()

	for _, e := range recent {
		k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
		if _, ok := ee.Events[k]; !ok {
			oid := catalog.NewEvent()
			device, door, card := lookup(e)
			ee.Events[k] = NewEvent(oid, e, device, door, card)
		}
	}
}

func validate(ee Events) error {
	return nil
}

func scrub(ee Events) error {
	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
