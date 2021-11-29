package types

import (
	"fmt"
	"testing"
	"time"
)

func TestDateTimeStringer(t *testing.T) {
	datetime := DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))

	tests := []struct {
		dt       *DateTime
		expected string
	}{
		{nil, ""},
		{&datetime, "2021-02-28 12:34:56"},
	}

	for _, v := range tests {
		s := fmt.Sprintf("%v", v.dt)

		if s != v.expected {
			t.Errorf("Invalid date/time string - expected:%v, got:%v", v.expected, s)
		}
	}
}

//func TestDateTimeStringer2(t *testing.T) {
//	datetime := DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 789, time.Local))
//
//	tests := []struct {
//		dt       DateTime
//		expected string
//	}{
//		{DateTime{}, "0001-01-01 00:00:00"},
//		{datetime, "2021-02-28 12:34:56"},
//	}
//
//	for _, v := range tests {
//		s := fmt.Sprintf("%v", v.dt)
//
//		if s != v.expected {
//			t.Errorf("Invalid date/time string - expected:%v, got:%v", v.expected, s)
//		}
//	}
//}
