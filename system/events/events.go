package events

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Events struct {
	events map[key]Event
	sync.RWMutex
}

type key struct {
	deviceID uint32
	index    uint32
}

type keyx struct {
	deviceID  uint32
	index     uint32
	timestamp time.Time
}

const EventsOID = schema.EventsOID
const EventsFirst = schema.EventsFirst
const EventsLast = schema.EventsLast

func newKeyX(deviceID uint32, index uint32, timestamp time.Time) keyx {
	year, month, day := timestamp.Date()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()
	location := timestamp.Location()
	t := time.Date(year, month, day, hour, minute, second, 0, location)

	return keyx{
		deviceID:  deviceID,
		index:     index,
		timestamp: t,
	}
}

func NewEvents() Events {
	return Events{
		events: map[key]Event{},
	}
}

func (ee *Events) AsObjects(start, max int, auth auth.OpAuth) []schema.Object {
	ee.RLock()
	defer ee.RUnlock()

	objects := []schema.Object{}
	keys := []keyx{}

	for _, e := range ee.events {
		keys = append(keys, newKeyX(e.DeviceID, e.Index, time.Time(e.Timestamp)))
	}

	sort.SliceStable(keys, func(i, j int) bool {
		p := keys[i]
		q := keys[j]
		return q.timestamp.Before(p.timestamp)
	})

	ix := start
	count := 0
	for ix < len(keys) && count < max {
		k := key{
			deviceID: keys[ix].deviceID,
			index:    keys[ix].index,
		}

		if e, ok := ee.events[k]; ok {
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
		k := key{
			deviceID: keys[0].deviceID,
			index:    keys[0].index,
		}

		l := key{
			keys[len(keys)-1].deviceID,
			keys[len(keys)-1].index,
		}

		if first, ok := ee.events[k]; ok {
			catalog.Join(&objects, catalog.NewObject2(EventsOID, EventsFirst, first.OID))
		}

		if last, ok := ee.events[l]; ok {
			catalog.Join(&objects, catalog.NewObject2(EventsOID, EventsLast, last.OID))
		}
	}

	return objects
}

// FIXME searches events list sequentially to find matching OID. But not ever actually invoked (as yet...)
func (ee *Events) UpdateByOID(auth auth.OpAuth, oid schema.OID, value string) ([]interface{}, error) {
	ee.Lock()
	defer ee.Unlock()

	var objects = []interface{}{}
	var err error

	for k, e := range ee.events {
		if e.OID.Contains(oid) {
			if objects, err = e.set(auth, oid, value); err == nil {
				ee.events[k] = e
			}
		}
	}

	return objects, err
}

func (ee *Events) Load(blob json.RawMessage) error {
	ee.Lock()
	defer ee.Unlock()

	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var e Event
		if err := e.deserialize(v); err == nil {
			deviceID := e.DeviceID

			k := key{
				deviceID: deviceID,
				index:    e.Index,
			}

			if u, ok := ee.events[k]; ok {
				return fmt.Errorf("%v  duplicate events (%v and %v)", k, e.OID, u.OID)
			} else {
				ee.events[k] = e
			}
		}
	}

	for _, e := range ee.events {
		catalog.PutT(e.CatalogEvent, e.OID)
	}

	return nil
}

func (ee *Events) Save() (json.RawMessage, error) {
	serializable := []json.RawMessage{}

	if err := ee.Validate(); err != nil {
		return nil, err
	}

	ee.RLock()
	defer ee.RUnlock()

	for _, e := range ee.events {
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (ee *Events) Print() {
	ee.RLock()
	defer ee.RUnlock()

	serializable := []json.RawMessage{}

	for _, e := range ee.events {
		if e.IsValid() && !e.IsDeleted() {
			if record, err := e.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- EVENTS\n%s\n", string(b))
	}
}

func (ee *Events) Clone() *Events {
	ee.RLock()
	defer ee.RUnlock()

	shadow := Events{
		events: map[key]Event{},
	}

	for k, e := range ee.events {
		shadow.events[k] = e.clone()
	}

	return &shadow
}

func (ee *Events) Validate() error {
	return nil
}

// NTS: original implementation - uses sequential search
// func (ee *Events) Missing(gaps int, controllers ...uint32) map[uint32][]types.Interval {
// 	missing := map[uint32][]types.Interval{}
//
// 	for _, c := range controllers {
// 		first := uint32(0)
// 		last := uint32(0)
//
// 		list := []uint32{}
// 		ee.events.Range(func(k, v any) bool {
// 			e := v.(uhppoted.Event)
// 			if e.DeviceID == c {
// 				list = append(list, e.Index)
// 			}
// 			return true
// 		})
//
// 		sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
//
// 		if N := len(list); N > 0 {
// 			first = list[0]
// 			last = list[N-1]
// 		}
//
// 		missing[c] = append(missing[c], types.Interval{From: last + 1, To: math.MaxUint32})
// 		if first > 1 {
// 			missing[c] = append(missing[c], types.Interval{From: 1, To: first - 1})
// 		}
//
// 		next := first
// 		ix := 0
// 		for ; ix < len(list) && gaps != 0; ix++ {
// 			v := list[ix]
// 			if v != next {
// 				from := next
// 				to := v - 1
// 				missing[c] = append(missing[c], types.Interval{From: from, To: to})
// 				gaps--
// 			}
// 			next = v + 1
// 		}
//
// 		if last > 0 && last > next && gaps != 0 {
// 			missing[c] = append(missing[c], types.Interval{From: next, To: last})
// 		}
// 	}
//
// 	return missing
// }

// NTS: first optimization - uses binary search
//      No perceived improvement in the benchmarks - looks like the copy from the Map
//      is the dominant cost. Cleaner algorithm though.
func (ee *Events) Missing(gaps int, controllers ...uint32) map[uint32][]types.Interval {
	ee.RLock()
	defer ee.RUnlock()

	missing := map[uint32][]types.Interval{}

	for _, c := range controllers {
		first := uint32(0)
		last := uint32(0)

		list := []uint32{}
		for _, e := range ee.events {
			if e.DeviceID == c {
				list = append(list, e.Index)
			}
		}

		sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })

		if N := len(list); N > 0 {
			first = list[0]
			last = list[N-1]
		}

		missing[c] = append(missing[c], types.Interval{From: last + 1, To: math.MaxUint32})
		if first > 1 {
			missing[c] = append(missing[c], types.Interval{From: 1, To: first - 1})
		}

		slice := list[0:]
		for len(slice) > 0 && gaps != 0 {
			ix := sort.Search(len(slice), func(i int) bool {
				return slice[i] != slice[0]+uint32(i)
			})

			if ix != len(slice) {
				from := slice[ix-1] + 1
				to := slice[ix] - 1
				missing[c] = append(missing[c], types.Interval{From: from, To: to})
				gaps--
			}

			slice = slice[ix:]
		}
	}

	return missing
}
func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	ee.Lock()
	ee.Unlock()

	for _, e := range recent {
		k := key{
			deviceID: e.DeviceID,
			index:    e.Index,
		}

		if _, ok := ee.events[k]; ok {
			continue
		}

		event := Event{
			CatalogEvent: catalog.CatalogEvent{
				DeviceID: e.DeviceID,
				Index:    e.Index,
			},
		}

		oid := catalog.NewT(event.CatalogEvent)
		device, door, card := lookup(e)

		ee.events[k] = NewEvent(oid, e, device, door, card)
	}
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
