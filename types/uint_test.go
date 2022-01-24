package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUint8String(t *testing.T) {
	tests := []struct {
		value    Uint8
		expected string
	}{
		{0, ""},
		{1, "1"},
		{255, "255"},
	}

	for _, v := range tests {
		s := fmt.Sprintf("%v", v.value)
		if s != v.expected {
			t.Errorf("%v: incorrect string value - expected:%v, got:%v", v.value, v.expected, s)
		}
	}
}

func TestUint8MarshalJSON(t *testing.T) {
	tests := []struct {
		value    Uint8
		expected string
	}{
		{0, `""`},
		{1, "1"},
		{255, "255"},
	}

	for _, v := range tests {
		if j, err := json.Marshal(v.value); err != nil {
			t.Errorf("Error marshalling %v (%v)", v.value, err)
		} else if string(j) != v.expected {
			t.Errorf("%v: incorrectly marshaled - expected:%v, got:%v", v.value, v.expected, string(j))
		}
	}
}

func TestUint32String(t *testing.T) {
	tests := []struct {
		value    Uint32
		expected string
	}{
		{0, ""},
		{1, "1"},
		{4294967295, "4294967295"},
	}

	for _, v := range tests {
		s := fmt.Sprintf("%v", v.value)
		if s != v.expected {
			t.Errorf("%v: incorrect string value - expected:%v, got:%v", v.value, v.expected, s)
		}
	}
}

func TestUint32MarshalJSON(t *testing.T) {
	tests := []struct {
		value    Uint32
		expected string
	}{
		{0, `""`},
		{1, "1"},
		{4294967295, "4294967295"},
	}

	for _, v := range tests {
		if j, err := json.Marshal(v.value); err != nil {
			t.Errorf("Error marshalling %v (%v)", v.value, err)
		} else if string(j) != v.expected {
			t.Errorf("%v: incorrectly marshaled - expected:%v, got:%v", v.value, v.expected, string(j))
		}
	}
}
