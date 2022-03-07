package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Events struct {
	events sync.Map
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

func NewEvents() Events {
	return Events{}
}

func (ee *Events) AsObjects(start, max int, auth auth.OpAuth) []catalog.Object {
	objects := []catalog.Object{}
	keys := []key{}

	ee.events.Range(func(k, v interface{}) bool {
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
		if v, ok := ee.events.Load(k); ok {
			e := v.(Event)
			if e.IsValid() || e.IsDeleted() {
				if l := e.AsObjects(auth); l != nil {
					objects = catalog.Join(objects, l...)
					count++
				}
			}
		}

		ix++
	}

	if len(keys) > 0 {
		first, _ := ee.events.Load(keys[0])
		last, _ := ee.events.Load(keys[len(keys)-1])

		if first != nil {
			objects = catalog.Join(objects, catalog.NewObject2(EventsOID, EventsFirst, first.(Event).OID))
		}

		if last != nil {
			objects = catalog.Join(objects, catalog.NewObject2(EventsOID, EventsLast, last.(Event).OID))
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

	ee.events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		if !e.OID.Contains(oid) {
			return true
		}

		if objects, err = e.set(auth, oid, value); err == nil {
			ee.events.Store(k, e)
		}

		return false
	})

	return objects, err
}

func (ee *Events) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var e Event
		if err := e.deserialize(v); err == nil {
			k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
			if x, ok := ee.events.Load(k); ok {
				return fmt.Errorf("%v  duplicate events (%v and %v)", k, e.OID, x.(Event).OID)
			} else {
				ee.events.Store(k, e)
			}
		}
	}

	ee.events.Range(func(k, v interface{}) bool {
		catalog.PutEvent(v.(Event).OID)
		return true
	})

	return nil
}

func (ee *Events) Save() (json.RawMessage, error) {
	if err := validate(ee); err != nil {
		return nil, err
	}

	if err := scrub(ee); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}

	ee.events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}

		return true
	})

	return json.MarshalIndent(serializable, "", "  ")
}

func (ee *Events) Print() {
	serializable := []json.RawMessage{}
	ee.events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}

		return true
	})

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- EVENTS\n%s\n", string(b))
	}
}

func (ee *Events) Clone() *Events {
	shadow := Events{}

	ee.events.Range(func(k, v interface{}) bool {
		shadow.events.Store(k, v.(Event).clone())
		return true
	})

	return &shadow
}

func (ee *Events) Validate() error {
	return nil
}

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	for _, e := range recent {
		k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
		if _, ok := ee.events.Load(k); !ok {
			oid := catalog.NewEvent()
			device, door, card := lookup(e)
			ee.events.Store(k, NewEvent(oid, e, device, door, card))
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
