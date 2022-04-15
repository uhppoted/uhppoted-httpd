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
	events sync.Map
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
	return Events{}
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
	missing := map[uint32][]types.Interval{}

	for _, c := range controllers {
		first := uint32(0)
		last := uint32(0)

		list := []uint32{}
		ee.events.Range(func(k, v any) bool {
			e := v.(Event)
			if e.DeviceID == c {
				list = append(list, e.Index)
			}
			return true
		})

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
	}
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
