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
	Events sync.Map `json:"events"`

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

func NewEvents() *Events {
	return &Events{}
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
			if x, ok := ee.Events.Load(k); ok {
				return fmt.Errorf("%v  duplicate events (%v and %v)", k, e.OID, x.(Event).OID)
			} else {
				ee.Events.Store(k, e)
			}
		}
	}

	ee.Events.Range(func(k, v interface{}) bool {
		catalog.PutEvent(v.(Event).OID)
		return true
	})

	ee.file = file

	return nil
}

func (ee *Events) Save() error {
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

	ee.Events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable.Events = append(serializable.Events, record)
			}
		}

		return true
	})

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

func (ee *Events) Print() {
	if b, err := json.MarshalIndent(&ee.Events, "", "  "); err == nil {
		fmt.Printf("----------------- EVENTS\n%s\n", string(b))
	}
}

func (ee *Events) Clone() *Events {
	shadow := Events{
		file: ee.file,
	}

	ee.Events.Range(func(k, v interface{}) bool {
		shadow.Events.Store(k, v.(Event).clone())
		return true
	})

	return &shadow
}

func (ee *Events) AsObjects(start, max int) []interface{} {
	objects := []interface{}{}
	keys := []key{}

	ee.Events.Range(func(k, v interface{}) bool {
		keys = append(keys, k.(key))
		return true
	})

	sort.SliceStable(keys, func(i, j int) bool {
		p := keys[i]
		q := keys[j]
		return q.timestamp.Before(p.timestamp)
	})

	ix := start
	count := 0
	for ix < len(keys) && count < max {
		k := keys[ix]
		if v, ok := ee.Events.Load(k); ok {
			e := v.(Event)
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
		first, _ := ee.Events.Load(keys[0])
		last, _ := ee.Events.Load(keys[len(keys)-1])

		if first != nil {
			objects = append(objects, catalog.NewObject2(EventsOID, EventsFirst, first.(Event).OID))
		}

		if last != nil {
			objects = append(objects, catalog.NewObject2(EventsOID, EventsLast, last.(Event).OID))
		}

	}

	return objects
}

func (ee *Events) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	if ee == nil {
		return nil, nil
	}

	var objects = []interface{}{}
	var err error

	ee.Events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		if !e.OID.Contains(oid) {
			return true
		}

		if objects, err = e.set(auth, oid, value); err == nil {
			ee.Events.Store(k, e)
		}

		return false
	})

	return objects, err
}

func (ee *Events) Validate() error {
	return nil
}

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	for _, e := range recent {
		k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
		if _, ok := ee.Events.Load(k); !ok {
			oid := catalog.NewEvent()
			device, door, card := lookup(e)
			ee.Events.Store(k, NewEvent(oid, e, device, door, card))
		}
	}
}

func validate(ee *Events) error {
	return nil
}

func scrub(ee *Events) error {
	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
