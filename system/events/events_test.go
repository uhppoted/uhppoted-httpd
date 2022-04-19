package events

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestEventsMissingWithNoGaps(t *testing.T) {
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1001); ix <= 1032; ix++ {
		k := eventKey{
			deviceID: 201020304,
			index:    ix}

		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				DeviceID: 201020304,
				Index:    ix}}

		events.events[k] = e
	}

	for ix := uint32(1); ix <= 69; ix++ {
		k := eventKey{
			deviceID: 405419896,
			index:    ix}

		e := Event{
			CatalogEvent: catalog.CatalogEvent{
				DeviceID: 405419896,
				Index:    ix}}

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
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1); ix <= 69; ix++ {
		if ix < 37 || ix > 43 {
			k := eventKey{
				deviceID: 405419896,
				index:    ix}

			e := Event{
				CatalogEvent: catalog.CatalogEvent{
					DeviceID: 405419896,
					Index:    ix}}

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
	events := Events{
		events: map[eventKey]Event{},
	}

	for ix := uint32(1); ix <= 69; ix++ {
		if !(ix >= 13 && ix <= 19) && !(ix >= 37 && ix <= 43) && !(ix >= 53 && ix <= 59) {
			k := eventKey{
				deviceID: 405419896,
				index:    ix}

			e := Event{
				CatalogEvent: catalog.CatalogEvent{
					DeviceID: 405419896,
					Index:    ix}}

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
			k := eventKey{
				deviceID: 405419896,
				index:    ix}

			e := Event{
				CatalogEvent: catalog.CatalogEvent{
					DeviceID: 405419896,
					Index:    ix}}

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
		events.Missing(-1, 405419896)
	}

	dt := time.Now().Sub(start).Milliseconds() / int64(b.N)
	if dt > 50 {
		b.Errorf("too slow (%vms measured over %v iterations)", dt, b.N)
	}
}
