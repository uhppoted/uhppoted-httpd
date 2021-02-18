package system

import (
	"reflect"
	"testing"
	"time"
)

func TestTimezone(t *testing.T) {
	local := time.Local
	utc, _ := time.LoadLocation("UTC")
	utc7n := time.FixedZone("UTC-7", -7*60*60)
	utc7p := time.FixedZone("UTC+7", +7*60*60)

	tests := []struct {
		String   string
		Expected *time.Location
	}{
		{"2021-02-18 12:26:23 PST", local},
		{"2021-02-18 12:26:23 UTC", utc},
		{"2021-02-18 12:26:23 UTC-7", utc7n},
		{"2021-02-18 12:26:23 -0700", utc7n},
		{"2021-02-18 12:26:23 -07:00", utc7n},
		{"2021-02-18 12:26:23 UTC+7", utc7p},
		{"2021-02-18 12:26:23 +0700", utc7p},
		{"2021-02-18 12:26:23 +07:00", utc7p},
		{"2021-02-18 12:26:23", local},
	}

	for _, v := range tests {
		tz, err := timezone(v.String)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(tz, v.Expected) {
			t.Errorf("%s: incorrect timezone - expected:%#v, got:%#v", v.String, v.Expected, tz)
		}
	}
}
