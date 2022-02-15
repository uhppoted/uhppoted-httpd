package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestTimestampNow(t *testing.T) {
	now := TimestampNow()
	time.Sleep(100 * time.Millisecond)
	then := TimestampNow()

	if time.Time(now).Nanosecond() != 0 {
		t.Errorf("TimestampNow returned non-zero nanoseconds (%v)", now)
	}

	if time.Time(then).Nanosecond() != 0 {
		t.Errorf("TimestampNow returned non-zero nanoseconds (%v)", then)
	}
}

func TestTimestampBeforeIgnoresSubseconds(t *testing.T) {
	timestamp := Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 345, time.Local))
	reference := time.Date(2021, time.February, 28, 12, 34, 56, 678, time.Local)

	if timestamp.Before(reference) {
		t.Errorf("Expected Timestamp.Before to ignore subsecond differences")
	}
}

func TestTimestampAdd(t *testing.T) {
	timestamp := Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))
	expected := Timestamp(time.Date(2021, time.February, 28, 15, 34, 56, 789, time.Local).Truncate(1 * time.Second))

	timestamp = timestamp.Add(3 * time.Hour)

	if timestamp != expected {
		t.Errorf("Incorrect date/time - expected:%v, got:%v", expected, timestamp)
	}

}

func TestTimestampMarshalJSON(t *testing.T) {
	utc := time.Date(2021, time.February, 28, 12, 34, 56, 789, time.UTC)
	local := time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local)
	zero := Timestamp{}

	tests := []struct {
		timestamp Timestamp
		expected  string
	}{
		{Timestamp(utc), `"2021-02-28 12:34:56 UTC"`},
		{Timestamp(local), local.Format(`"2006-01-02 15:04:05 MST"`)},
		{zero, `""`},
	}

	for _, v := range tests {
		if b, err := json.Marshal(v.timestamp); err != nil {
			t.Errorf("Error marshalling %v (%v)", v.timestamp, err)
		} else if string(b) != v.expected {
			t.Errorf("Timestamp %v incorrectly marshalled - expected:%v, got:%v", v.timestamp, v.expected, string(b))
		}
	}
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	utc := Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.UTC))
	local := Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))
	zero := Timestamp{}

	tests := []struct {
		json     string
		expected Timestamp
	}{
		{`"2021-02-28 12:34:56 UTC"`, utc},
		{`"2021-02-28 12:34:56"`, local},
		{`""`, zero},
	}

	for _, v := range tests {
		var dt Timestamp

		if err := json.Unmarshal([]byte(v.json), &dt); err != nil {
			t.Errorf("Error unmarshalling %v (%v)", v.json, err)
		} else {
			p := fmt.Sprintf("%v", &dt)
			q := fmt.Sprintf("%v", &v.expected)

			if p != q {
				t.Errorf("Invalid date/time - expected:%v, got:%v", q, p)
			}
		}
	}
}

func TestTimestampString(t *testing.T) {
	var tz, _ = time.LoadLocation("UTC")
	var zero = Timestamp{}
	var timestamp = Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 789, tz))

	tests := []struct {
		dt       interface{}
		expected string
	}{
		{timestamp, "2021-02-28 12:34:56 UTC"},
		{&timestamp, "2021-02-28 12:34:56 UTC"},
		{zero, ""},
		{&zero, ""},
	}

	for _, v := range tests {
		s := fmt.Sprintf("%v", v.dt)

		if s != v.expected {
			t.Errorf("Invalid date/time string - expected:%v, got:%v", v.expected, s)
		}
	}
}
