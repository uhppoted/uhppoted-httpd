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
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Events struct {
	events sync.Map
	first  map[uint32]uint32
	last   map[uint32]uint32
}

type key struct {
	deviceID  uint32
	index     uint32
	timestamp time.Time
}

const EventsOID = schema.EventsOID
const EventsFirst = schema.EventsFirst
const EventsLast = schema.EventsLast

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
		first: map[uint32]uint32{},
		last:  map[uint32]uint32{},
	}
}

func (ee *Events) AsObjects(start, max int, auth auth.OpAuth) []schema.Object {
	objects := []schema.Object{}
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
					catalog.Join(&objects, l...)
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
			catalog.Join(&objects, catalog.NewObject2(EventsOID, EventsFirst, first.(Event).OID))
		}

		if last != nil {
			catalog.Join(&objects, catalog.NewObject2(EventsOID, EventsLast, last.(Event).OID))
		}
	}

	return objects
}

func (ee *Events) UpdateByOID(auth auth.OpAuth, oid schema.OID, value string) ([]interface{}, error) {
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
			deviceID := e.DeviceID

			if f, ok := ee.first[deviceID]; !ok || f > e.Index {
				ee.first[deviceID] = e.Index
			}

			if l, ok := ee.last[deviceID]; !ok || l < e.Index {
				ee.last[deviceID] = e.Index
			}

			k := newKey(deviceID, e.Index, time.Time(e.Timestamp))
			if x, ok := ee.events.Load(k); ok {
				return fmt.Errorf("%v  duplicate events (%v and %v)", k, e.OID, x.(Event).OID)
			} else {
				ee.events.Store(k, e)
			}
		}
	}

	ee.events.Range(func(k, v interface{}) bool {
		e := v.(Event)
		catalog.PutT(e.CatalogEvent, e.OID)
		return true
	})

	return nil
}

func (ee *Events) Save() (json.RawMessage, error) {
	if err := ee.Validate(); err != nil {
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

func (ee Events) Indices(deviceID uint32) (first uint32, last uint32) {
	first, _ = ee.first[deviceID]
	last, _ = ee.last[deviceID]

	return
}

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	for _, e := range recent {
		k := newKey(e.DeviceID, e.Index, time.Time(e.Timestamp))
		if _, ok := ee.events.Load(k); ok {
			continue
		}

		oid := catalog.NewT(Event{}.CatalogEvent)
		if _, ok := ee.events.Load(oid); ok {
			warn(fmt.Errorf("catalog returned duplicate OID (%v)", oid))
			continue
		}

		device, door, card := lookup(e)
		ee.events.Store(k, NewEvent(oid, e, device, door, card))

		if ix, ok := ee.first[deviceID]; !ok || e.Index < ix {
			ee.first[deviceID] = e.Index
		}

		if ix, ok := ee.last[deviceID]; !ok || e.Index > ix {
			ee.last[deviceID] = e.Index
		}
	}
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
