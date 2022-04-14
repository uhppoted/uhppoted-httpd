package events

import (
	"math"
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-lib/uhppoted"

	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestEventsMissingWithNoGaps(t *testing.T) {
	events := Events{}

	for ix := uint32(1001); ix <= 1032; ix++ {
		events.events.Store(201020304+ix, uhppoted.Event{
			Index:    ix,
			DeviceID: 201020304,
		})
	}

	for ix := uint32(1); ix <= 69; ix++ {
		events.events.Store(405419896+ix, uhppoted.Event{
			Index:    ix,
			DeviceID: 405419896,
		})
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
	events := Events{}

	for ix := uint32(1); ix <= 69; ix++ {
		if ix < 37 || ix > 43 {
			events.events.Store(405419896+ix, uhppoted.Event{
				Index:    ix,
				DeviceID: 405419896,
			})
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
	events := Events{}

	for ix := uint32(1); ix <= 69; ix++ {
		if !(ix >= 13 && ix <= 19) && !(ix >= 37 && ix <= 43) && !(ix >= 53 && ix <= 59) {
			events.events.Store(405419896+ix, uhppoted.Event{
				Index:    ix,
				DeviceID: 405419896,
			})
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
	events := Events{}

	for ix := uint32(1); ix <= 69; ix++ {
		if !(ix <= 5) &&
			!(ix >= 13 && ix <= 15) &&
			!(ix >= 23 && ix <= 27) &&
			!(ix >= 37 && ix <= 39) &&
			!(ix >= 44 && ix <= 48) &&
			!(ix >= 57 && ix <= 61) {
			events.events.Store(405419896+ix, uhppoted.Event{
				Index:    ix,
				DeviceID: 405419896,
			})
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
	lots := Events{}

	for ix := uint32(1); ix <= 100000; ix++ {
		lots.events.Store(405419896+ix, uhppoted.Event{
			Index:    ix,
			DeviceID: 405419896,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lots.Missing(-1, 405419896)
	}
}
