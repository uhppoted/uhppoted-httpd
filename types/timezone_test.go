package types

import (
	"testing"
	"time"
)

// Ref. https://github.com/golang/go/issues/12388
func TestTimezone(t *testing.T) {
	local := time.Local
	pst := time.FixedZone("PST", +8*60*60) // NOTE: this is potentially problematic - there doesn't seem to be a way to reliably get a timezone by short code
	mst, _ := time.LoadLocation("MST")
	utc, _ := time.LoadLocation("UTC")
	utc7p := time.FixedZone("UTC+7", +7*60*60)
	utc7n := time.FixedZone("UTC-7", -7*60*60)

	tests := []struct {
		String   string
		Expected *time.Location
	}{
		{"", local},
		{"2021-02-18 12:26:23 PST", pst},
		{"2021-02-18 12:26:23 MST", mst},
		{"2021-02-18 12:26:23 UTC", utc},
		{"2021-02-18 12:26:23 UTC+7", utc7p},
		{"2021-02-18 12:26:23 +0700", utc7p},
		{"2021-02-18 12:26:23 +07:00", utc7p},
		{"2021-02-18 12:26:23 UTC-7", utc7n},
		{"2021-02-18 12:26:23 -0700", utc7n},
		{"2021-02-18 12:26:23 -07:00", utc7n},
		{"2021-02-18 12:26:23", local},
		{"UTC", utc},
		{"UTC+0", utc},
		{"UTC-0", utc},
		{"UTC+7", utc7p},
		{"UTC-7", utc7n},
	}

	for _, v := range tests {
		tz, err := Timezone(v.String)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tz.String() != v.Expected.String() {
			t.Errorf("%s: incorrect timezone - expected:%v, got:%v", v.String, v.Expected, tz)
		}
	}
}
