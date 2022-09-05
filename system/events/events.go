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
	events map[eventKey]Event
	sync.RWMutex
}

type eventKey struct {
	deviceID uint32
	index    uint32
}

type objectKey struct {
	deviceID  uint32
	index     uint32
	timestamp time.Time
}

const EventsOID = schema.EventsOID
const EventsFirst = schema.EventsFirst
const EventsLast = schema.EventsLast

var cache = struct {
	events struct {
		events map[uint32][]uint32
		dirty  bool
	}
	objects struct {
		objects []objectKey
		dirty   bool
	}
}{
	events: struct {
		events map[uint32][]uint32
		dirty  bool
	}{
		events: map[uint32][]uint32{},
		dirty:  true,
	},
	objects: struct {
		objects []objectKey
		dirty   bool
	}{
		objects: []objectKey{},
		dirty:   true,
	},
}

func NewEvents() Events {
	return Events{
		events: map[eventKey]Event{},
	}
}

func (ee *Events) AsObjects(start, max int, auth auth.OpAuth) []schema.Object {
	ee.RLock()
	defer ee.RUnlock()

	objects := []schema.Object{}
	keys := cache.objects.objects

	if cache.objects.dirty {
		keys = []objectKey{}
		for k, e := range ee.events {
			keys = append(keys, objectKey{
				deviceID:  k.deviceID,
				index:     k.index,
				timestamp: time.Time(e.Timestamp).Round(time.Second),
			})
		}

		sort.SliceStable(keys, func(i, j int) bool {
			return keys[j].timestamp.Before(keys[i].timestamp)
		})

		cache.objects.objects = keys
		cache.objects.dirty = false
	}

	ix := start
	count := 0
	for ix < len(keys) && count < max {
		k := eventKey{
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
		k := eventKey{keys[len(keys)-1].deviceID, keys[len(keys)-1].index}
		l := eventKey{deviceID: keys[0].deviceID, index: keys[0].index}

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

			k := eventKey{
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

	cache.events.dirty = true
	cache.objects.dirty = true

	for _, e := range ee.events {
		catalog.PutT(e.CatalogEvent)
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
		events: map[eventKey]Event{},
	}

	for k, e := range ee.events {
		shadow.events[k] = e.clone()
	}

	return &shadow
}

func (ee *Events) Validate() error {
	return nil
}

// NTS: for 1000000 events (i.e. in the expected range), binary search improves only slightly
//
//	on a sequential linear traverse but it's a cleaner algorithm and scales somewhat better.
//	Copying from the map is the dominant cost.
func (ee *Events) Missing(gaps int, controllers ...uint32) map[uint32][]types.Interval {
	ee.RLock()
	defer ee.RUnlock()

	missing := map[uint32][]types.Interval{}

	if cache.events.dirty {
		m := map[uint32][]uint32{}
		for _, e := range ee.events {
			var k = e.DeviceID
			var l = m[k]

			m[k] = append(l, e.Index)
		}

		for k, list := range m {
			sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
			m[k] = list
		}

		cache.events.dirty = false
		cache.events.events = m
	}

	for _, c := range controllers {
		first := uint32(0)
		last := uint32(0)
		list := cache.events.events[c]

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
		k := eventKey{
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

	cache.events.dirty = true
	cache.objects.dirty = true
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
