package events

import (
	"math"
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestEventsMissingWithNoGaps(t *testing.T) {
	events := Events{
		first: map[uint32]uint32{
			201020304: 1001,
			303986753: 0,
			405419896: 1,
		},
		last: map[uint32]uint32{
			201020304: 1032,
			303986753: 0,
			405419896: 69,
		},
	}

	expected := map[uint32][]types.Interval{
		101020304: {
			{From: 1, To: math.MaxUint32},
		},
		201020304: {
			{From: 1, To: 1000}, {From: 1033, To: math.MaxUint32},
		},
		303986753: {
			{From: 1, To: math.MaxUint32},
		},
		405419896: {
			{From: 70, To: math.MaxUint32},
		},
	}

	missing := events.Missing(101020304, 201020304, 303986753, 405419896)

	if !reflect.DeepEqual(missing, expected) {
		t.Errorf("Incorrect missing events list\n   expected:%v\n   got:     %v", expected, missing)
	}
}
