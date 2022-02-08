package auth

import (
	"testing"
)

func TestAuthorizatorUID(t *testing.T) {
	tests := []struct {
		a        *authorizator
		expected string
	}{
		{&authorizator{uid: "qwerty"}, "qwerty"},
		{nil, "?"},
	}

	for _, test := range tests {
		if uid := test.a.UID(); uid != test.expected {
			t.Errorf("Incorrect UID value - expected:%v, got:%v", test.expected, uid)
		}
	}
}
