package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestDateTimeString(t *testing.T) {
	datetime := DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))

	tests := []struct {
		dt       DateTime
		expected string
	}{
		{datetime, "2021-02-28 12:34:56"},
	}

	for _, v := range tests {
		s := fmt.Sprintf("%v", v.dt)

		if s != v.expected {
			t.Errorf("Invalid date/time string - expected:%v, got:%v", v.expected, s)
		}
	}
}

func TestDateTimeUnmarshalJSON(t *testing.T) {
	utc := DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.UTC))
	local := DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))

	tests := []struct {
		json     string
		expected DateTime
	}{
		{`"2021-02-28 12:34:56 UTC"`, utc},
		{`"2021-02-28 12:34:56"`, local},
	}

	for _, v := range tests {
		var dt DateTime

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
