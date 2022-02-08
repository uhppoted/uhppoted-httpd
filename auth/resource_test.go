package auth

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
)

func TestResourceMarshalJSON(t *testing.T) {
	tests := []struct {
		r        resource
		expected string
	}{
		{
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("admin"),
			},
			`{"path":"^[a-z]+[0-9]+$","authorised":"admin"}`,
		},
	}

	for _, test := range tests {
		if bytes, err := json.Marshal(test.r); err != nil {
			t.Errorf("Unexpected error marshalling auth resource (%v)", err)
		} else if string(bytes) != test.expected {
			t.Errorf("Incorrectly serialized:\n   expected:%v\n   got:     %v", test.expected, string(bytes))
		}
	}
}

func TestResourceUnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected resource
	}{
		{
			`{"path":"[a-z]+[0-9]+","authorised":"admin"}`,
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("admin"),
			},
		},
		{
			`{"path":"[a-z]+[0-9]+$","authorised":"admin"}`,
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("admin"),
			},
		},
		{
			`{"path":"^[a-z]+[0-9]+","authorised":"admin"}`,
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("admin"),
			},
		},
		{
			`{"path":"^[a-z]+[0-9]+$","authorised":"admin"}`,
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("admin"),
			},
		},
	}

	for _, test := range tests {
		r := resource{}
		if err := json.Unmarshal([]byte(test.json), &r); err != nil {
			t.Errorf("Unexpected error marshalling auth resource (%v)", err)
		} else if !reflect.DeepEqual(r, test.expected) {
			t.Errorf("Incorrectly deserialized:\n   expected:%v\n   got:     %v", test.expected, r)
		}
	}
}
