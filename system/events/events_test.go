package events

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestEventsAsObjects(t *testing.T) {
	cache.objects.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	oid := schema.EventsOID
	base := time.Date(2022, time.April, 20, 12, 34, 50, 0, time.UTC)
	delta := 5 * time.Minute

	for i := 0; i < 32; i++ {
		ix := uint32(1001 + i)
		k := eventKey{deviceID: 201020304, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 201020304,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(base.Add(time.Duration(i) * delta)),
		}

		events.events[k] = e
	}

	expected := []schema.Object{
		schema.Object{OID: "0.6.1020.1", Value: "2022-04-20 14:09:50"},
		schema.Object{OID: "0.6.1019.1", Value: "2022-04-20 14:04:50"},
		schema.Object{OID: "0.6.1018.1", Value: "2022-04-20 13:59:50"},
		schema.Object{OID: "0.6.1017.1", Value: "2022-04-20 13:54:50"},
		schema.Object{OID: "0.6.1016.1", Value: "2022-04-20 13:49:50"},
	}

	objects := events.AsObjects(12, 5, nil)

	// Check first/last
	for _, o := range objects {
		if o.OID == "0.6.0.1" {
			if fmt.Sprintf("%v", o.Value) != "0.6.1001" {
				t.Errorf("Incorrect AsObjects list 'first' event\n   expected:%v\n   got:     %#v", "0.6.1001", o.Value)
			}
		}

		if o.OID == "0.6.0.2" {
			if fmt.Sprintf("%v", o.Value) != "0.6.1032" {
				t.Errorf("Incorrect AsObjects list 'last' event\n   expected:%v\n   got:     %#v", "0.6.1032", o.Value)
			}
		}
	}

	// Check timestamps
	timestamps := []schema.Object{}
	for _, o := range objects {
		if o.OID.HasSuffix(".1") && o.OID != "0.6.0.1" {
			timestamps = append(timestamps, schema.Object{OID: o.OID, Value: fmt.Sprintf("%v", o.Value)})
		}
	}

	if !reflect.DeepEqual(timestamps, expected) {
		t.Errorf("Incorrect AsObjects list\n   expected:%v\n   got:     %v", expected, timestamps)
	}
}

func TestEventsAsObjectsWithMultipleDevices(t *testing.T) {
	cache.objects.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	oid := schema.EventsOID
	base := time.Date(2022, time.April, 20, 12, 34, 50, 0, time.UTC)
	delta := 5 * time.Minute
	for i := 0; i < 32; i++ {
		ix := uint32(1001 + i)
		timestamp := base.Add(time.Duration(i) * delta)
		k := eventKey{deviceID: 201020304, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 201020304,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(timestamp),
		}

		events.events[k] = e
	}

	delta = 5 * time.Minute
	for i := 0; i < 32; i++ {
		ix := uint32(1 + i)
		timestamp := base.Add(time.Duration(i)*delta + time.Minute)
		k := eventKey{deviceID: 405419896, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 405419896,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(timestamp),
		}

		events.events[k] = e
	}

	expected := []schema.Object{
		schema.Object{OID: "0.6.22.1", Value: "2022-04-20 14:20:50"},
		schema.Object{OID: "0.6.1022.1", Value: "2022-04-20 14:19:50"},
		schema.Object{OID: "0.6.21.1", Value: "2022-04-20 14:15:50"},
		schema.Object{OID: "0.6.1021.1", Value: "2022-04-20 14:14:50"},
		schema.Object{OID: "0.6.20.1", Value: "2022-04-20 14:10:50"},
	}

	objects := events.AsObjects(20, 5, nil)

	// Check first/last
	for _, o := range objects {
		if o.OID == "0.6.0.1" {
			if fmt.Sprintf("%v", o.Value) != "0.6.1001" {
				t.Errorf("Incorrect AsObjects list 'first' event\n   expected:%v\n   got:     %#v", "0.6.1001", o.Value)
			}
		}

		if o.OID == "0.6.0.2" {
			if fmt.Sprintf("%v", o.Value) != "0.6.32" {
				t.Errorf("Incorrect AsObjects list 'first' event\n   expected:%v\n   got:     %#v", "0.6.32", o.Value)
			}
		}
	}

	// Check timestamps
	timestamps := []schema.Object{}
	for _, o := range objects {
		if o.OID.HasSuffix(".1") && o.OID != "0.6.0.1" {
			timestamps = append(timestamps, schema.Object{OID: o.OID, Value: fmt.Sprintf("%v", o.Value)})
		}
	}

	if !reflect.DeepEqual(timestamps, expected) {
		t.Errorf("Incorrect AsObjects list\n   expected:%#v\n   got:     %#v", expected, timestamps)
	}
}

func TestEventsAsObjectsFromZero(t *testing.T) {
	cache.objects.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	oid := schema.EventsOID
	base := time.Date(2022, time.April, 20, 12, 34, 50, 0, time.UTC)
	delta := 5 * time.Minute

	for i := 0; i < 32; i++ {
		ix := uint32(1001 + i)
		k := eventKey{deviceID: 201020304, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 201020304,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(base.Add(time.Duration(i) * delta)),
		}

		events.events[k] = e
	}

	expected := []schema.Object{
		schema.Object{OID: "0.6.1032.1", Value: "2022-04-20 15:09:50"},
		schema.Object{OID: "0.6.1031.1", Value: "2022-04-20 15:04:50"},
		schema.Object{OID: "0.6.1030.1", Value: "2022-04-20 14:59:50"},
	}

	objects := events.AsObjects(0, 3, nil)

	// Check timestamps
	timestamps := []schema.Object{}
	for _, o := range objects {
		if o.OID.HasSuffix(".1") && o.OID != "0.6.0.1" {
			timestamps = append(timestamps, schema.Object{OID: o.OID, Value: fmt.Sprintf("%v", o.Value)})
		}
	}

	if !reflect.DeepEqual(timestamps, expected) {
		t.Errorf("Incorrect AsObjects list\n   expected:%#v\n   got:     %#v", expected, timestamps)
	}
}

func TestEventsMissingWithNoGaps(t *testing.T) {
	cache.events.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1001); ix <= 1032; ix++ {
		k := eventKey{deviceID: 201020304, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{DeviceID: 201020304, Index: ix},
		}

		events.events[k] = e
	}

	for ix := uint32(1); ix <= 69; ix++ {
		k := eventKey{deviceID: 405419896, index: ix}
		e := Event{
			CatalogEvent: catalog.CatalogEvent{DeviceID: 405419896, Index: ix},
		}

		events.events[k] = e
	}

	expected := map[uint32][]types.Interval{
		201020304: {
			{From: 1033, To: math.MaxUint32}, {From: 1, To: 1000},
		},
		303986753: {
			{From: 1, To: math.MaxUint32},
		},
		405419896: {
			{From: 70, To: math.MaxUint32},
		},
	}

	missing := events.Missing(-1, 201020304, 303986753, 405419896)

	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}
}

func TestEventsMissingWithGaps(t *testing.T) {
	cache.events.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1); ix <= 69; ix++ {
		if ix < 37 || ix > 43 {
			k := eventKey{deviceID: 405419896, index: ix}
			e := Event{
				CatalogEvent: catalog.CatalogEvent{DeviceID: 405419896, Index: ix},
			}

			events.events[k] = e
		}
	}

	expected := map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 37, To: 43},
		},
	}

	missing := events.Missing(-1, 405419896)

	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}
}

func TestEventsMissingWithMultipleGaps(t *testing.T) {
	cache.events.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1); ix <= 69; ix++ {
		if !(ix >= 13 && ix <= 19) && !(ix >= 37 && ix <= 43) && !(ix >= 53 && ix <= 59) {
			k := eventKey{deviceID: 405419896, index: ix}
			e := Event{
				CatalogEvent: catalog.CatalogEvent{DeviceID: 405419896, Index: ix},
			}

			events.events[k] = e
		}
	}

	expected := map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 13, To: 19},
			{From: 37, To: 43},
			{From: 53, To: 59},
		},
	}

	missing := events.Missing(-1, 405419896)

	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}
}

func TestEventsMissingWithGapsLimit(t *testing.T) {
	cache.events.dirty = true
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1); ix <= 69; ix++ {
		if !(ix <= 5) &&
			!(ix >= 13 && ix <= 15) &&
			!(ix >= 23 && ix <= 27) &&
			!(ix >= 37 && ix <= 39) &&
			!(ix >= 44 && ix <= 48) &&
			!(ix >= 57 && ix <= 61) {
			k := eventKey{deviceID: 405419896, index: ix}
			e := Event{
				CatalogEvent: catalog.CatalogEvent{DeviceID: 405419896, Index: ix},
			}

			events.events[k] = e
		}
	}

	expected := map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 1, To: 5},
			{From: 13, To: 15},
			{From: 23, To: 27},
			{From: 37, To: 39},
			{From: 44, To: 48},
			{From: 57, To: 61},
		},
	}

	missing := events.Missing(-1, 405419896)
	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}

	expected = map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 1, To: 5},
		},
	}

	missing = events.Missing(0, 405419896)
	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}

	expected = map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 1, To: 5},
			{From: 13, To: 15},
		},
	}

	missing = events.Missing(1, 405419896)
	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}

	expected = map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 1, To: 5},
			{From: 13, To: 15},
			{From: 23, To: 27},
		},
	}

	missing = events.Missing(2, 405419896)
	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}

	expected = map[uint32][]types.Interval{
		405419896: {
			{From: 70, To: math.MaxUint32},
			{From: 1, To: 5},
			{From: 13, To: 15},
			{From: 23, To: 27},
			{From: 37, To: 39},
			{From: 44, To: 48},
			{From: 57, To: 61},
		},
	}

	missing = events.Missing(1000, 405419896)
	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}
}

func BenchmarkMissingEvents(b *testing.B) {
	list := []Event{}
	for ix := uint32(1); ix <= 100000; ix++ {
		if (ix >= 20001 && ix <= 20015) ||
			(ix >= 32101 && ix <= 32111) {
			continue
		}

		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				DeviceID: 405419896,
				Index:    ix}}

		list = append(list, e)
	}

	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	events := Events{
		events: map[eventKey]Event{},
	}

	for _, e := range list {
		k := eventKey{
			e.DeviceID,
			e.Index,
		}

		events.events[k] = e
	}

	b.ResetTimer()

	start := time.Now()
	for i := 0; i < b.N; i++ {
		cache.events.dirty = true
		events.Missing(-1, 405419896)
	}

	dt := time.Now().Sub(start).Milliseconds() / int64(b.N)
	if dt > 50 {
		b.Errorf("too slow (%vms measured over %v iterations)", dt, b.N)
	}
}

func BenchmarkMissingEventsWithCache(b *testing.B) {
	list := []Event{}
	for ix := uint32(1); ix <= 100000; ix++ {
		if (ix >= 20001 && ix <= 20015) ||
			(ix >= 32101 && ix <= 32111) {
			continue
		}

		e := Event{
			CatalogEvent: catalog.CatalogEvent{DeviceID: 405419896, Index: ix},
		}

		list = append(list, e)
	}

	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	events := Events{
		events: map[eventKey]Event{},
	}

	for _, e := range list {
		k := eventKey{e.DeviceID, e.Index}

		events.events[k] = e
	}

	events.Missing(-1, 405419896)

	b.ResetTimer()

	start := time.Now()
	for i := 0; i < b.N; i++ {
		events.Missing(-1, 405419896)
	}

	dt := time.Now().Sub(start).Milliseconds() / int64(b.N)
	if dt > 5 {
		b.Errorf("too slow (%vms measured over %v iterations)", dt, b.N)
	}
}

func BenchmarkEventsAsObjectsWithoutCaching(b *testing.B) {
	events := Events{
		events: map[eventKey]Event{},
	}

	oid := schema.EventsOID
	base := time.Date(2022, time.April, 20, 12, 34, 50, 0, time.UTC)
	delta := 5 * time.Minute
	list := []Event{}

	for i := 0; i < 100000; i++ {
		ix := uint32(1001 + i)
		dt := time.Duration(i+rand.Intn(50)-25) * delta

		list = append(list, Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 201020304,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(base.Add(dt)),
		})
	}

	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	for _, e := range list {
		k := eventKey{deviceID: e.DeviceID, index: e.Index}
		events.events[k] = e
	}

	b.ResetTimer()

	start := time.Now()
	for i := 0; i < b.N; i++ {
		cache.objects.dirty = true
		events.AsObjects(25423, 15, nil)
	}

	dt := time.Now().Sub(start).Milliseconds() / int64(b.N)
	if dt > 200 {
		b.Errorf("too slow (%vms measured over %v iterations)", dt, b.N)
	}
}

func BenchmarkEventsAsObjectsWithCaching(b *testing.B) {
	events := Events{
		events: map[eventKey]Event{},
	}

	oid := schema.EventsOID
	base := time.Date(2022, time.April, 20, 12, 34, 50, 0, time.UTC)
	delta := 5 * time.Minute
	list := []Event{}

	for i := 0; i < 100000; i++ {
		ix := uint32(1001 + i)
		dt := time.Duration(i+rand.Intn(50)-25) * delta

		list = append(list, Event{
			CatalogEvent: catalog.CatalogEvent{
				OID:      oid.AppendS(strconv.Itoa(int(ix))),
				DeviceID: 201020304,
				Index:    uint32(ix),
			},
			Timestamp: core.DateTime(base.Add(dt)),
		})
	}

	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	for _, e := range list {
		k := eventKey{deviceID: e.DeviceID, index: e.Index}
		events.events[k] = e
	}

	events.AsObjects(25423, 15, nil)

	b.ResetTimer()

	start := time.Now()
	for i := 0; i < b.N; i++ {
		events.AsObjects(25423, 15, nil)
	}

	dt := time.Now().Sub(start).Milliseconds() / int64(b.N)
	if dt > 5 {
		b.Errorf("too slow (%vms measured over %v iterations)", dt, b.N)
	}
}
