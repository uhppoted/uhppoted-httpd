package auth

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
)

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
