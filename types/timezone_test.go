package types

import (
	"testing"
	"time"
)

// Ref. https://github.com/golang/go/issues/12388
func TestTimezone(t *testing.T) {
	local := time.Local
	pst := time.FixedZone("PST", +8*60*60) // NOTE: this is potentially problematic - there doesn't seem to be a way to reliably get a timezone by short code
	pdt := time.FixedZone("PDT", +7*60*60) // NOTE: ditto
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
		{"2021-02-18 12:26:23 PDT", pdt},
		{"2021-02-18 12:26:23 MST", mst},
		{"2021-02-18 12:26:23 UTC", utc},
		{"2021-02-18 12:26:23 UTC+7", utc7p},
		{"2021-02-18 12:26:23 +0700", utc7p},
		{"2021-02-18 12:26:23 +07:00", utc7p},
		{"2021-02-18 12:26:23 UTC-7", utc7n},
		{"2021-02-18 12:26:23 -0700", utc7n},
		{"2021-02-18 12:26:23 -07:00", utc7n},
		{"2021-02-18 12:26:23", local},
		{"2021-02-18 12:26:23 pst", pst},
		{"2021-02-18 12:26:23 utc-7", utc7n},

		{"2021-02-18 12:26 PST", pst},
		{"2021-02-18 12:26 PDT", pdt},
		{"2021-02-18 12:26 MST", mst},
		{"2021-02-18 12:26 UTC", utc},
		{"2021-02-18 12:26 UTC+7", utc7p},
		{"2021-02-18 12:26 +0700", utc7p},
		{"2021-02-18 12:26 +07:00", utc7p},
		{"2021-02-18 12:26 UTC-7", utc7n},
		{"2021-02-18 12:26 -0700", utc7n},
		{"2021-02-18 12:26 -07:00", utc7n},
		{"2021-02-18 12:26", local},

		{"PST", pst},
		{"PDT", pdt},

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

func TestTimezoneAfricaCairo(t *testing.T) {
	expected := "Africa/Cairo"

	for _, s := range []string{"2022-04-27 20:05:33 Africa/Cairo", "Africa/Cairo"} {
		tz, err := Timezone(s)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tz.String() != expected {
			t.Errorf("%s: incorrect timezone - expected:%v, got:%v", s, expected, tz)
		}
	}
}
