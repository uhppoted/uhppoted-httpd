package auth

import (
	"reflect"
	"regexp"
	"testing"
)

func TestLocalSerialize(t *testing.T) {
	l := Local{
		private: private{
			keys: [][]byte{[]byte("qwerty")},
			users: map[string]*user{
				"hagrid": &user{
					Salt:     salt([]byte{0xef, 0xcd, 0x34, 0x12}),
					Password: "dragon",
					Role:     "grounds keeper",
				},
			},
			resources: []resource{
				resource{
					Path:       regexp.MustCompile("[a-z]+[0-9]+"),
					Authorised: regexp.MustCompile("[0-9]+[a-z]*"),
				},
			},
		},
	}

	expected := `{
  "users": {
    "hagrid": {
      "salt": "efcd3412",
      "password": "dragon",
      "role": "grounds keeper"
    }
  },
  "resources": [
    {
      "path": "[a-z]+[0-9]+",
      "authorised": "[0-9]+[a-z]*"
    }
  ]
}`

	bytes, err := l.serialize()
	if err != nil {
		t.Fatalf("Unexpected error marshalling auth.Local (%v)", err)
	}

	if string(bytes) != expected {
		t.Errorf("Incorrectly serialized:\n   expected:%v\n   got:     %v", expected, string(bytes))
	}
}

func TestLocalDeserialize(t *testing.T) {
	json := `{
  "users": {
    "hagrid": {
      "salt": "efcd3412",
      "password": "dragon",
      "role": "grounds keeper"
    }
  },
  "resources": [
    {
      "path": "[a-z]+[0-9]+",
      "authorised": "[0-9]+[a-z]*"
    }
  ]
}`

	expected := Local{
		private: private{
			keys: [][]byte{[]byte("qwerty")},
			users: map[string]*user{
				"hagrid": &user{
					Salt:     salt([]byte{0xef, 0xcd, 0x34, 0x12}),
					Password: "dragon",
					Role:     "grounds keeper",
				},
			},
			resources: []resource{
				resource{
					Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
					Authorised: regexp.MustCompile("[0-9]+[a-z]*"),
				},
			},
		},
	}

	local := Local{
		private: private{
			keys: [][]byte{[]byte("qwerty")},
		},
	}

	if err := local.deserialize([]byte(json)); err != nil {
		t.Fatalf("Unexpected error unmarshalling auth.Local (%v)", err)
	}

	// ... check this way because reflect.DeepEqual copies guard value
	if !reflect.DeepEqual(local.private.keys, expected.private.keys) {
		t.Errorf("Incorrectly deserialized:\n   expected:%x\n   got:     %x", expected.private.keys, local.private.keys)
	}

	if !reflect.DeepEqual(local.private.users, expected.private.users) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}

	if !reflect.DeepEqual(local.private.resources, expected.private.resources) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}
}

func TestLocalCopyKey(t *testing.T) {
	tests := []struct {
		key      []byte
		expected []byte
	}{
		{
			key: []byte(""),
			expected: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			key: []byte("qwerty"),
			expected: []byte{
				0x71, 0x77, 0x65, 0x72, 0x74, 0x79, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			key: []byte("abcdefghijklmnopqrstuvwxyz123456"),
			expected: []byte{
				0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70,
				0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36},
		},
		{
			key: []byte("abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJK"),
			expected: []byte{
				0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70,
				0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36},
		},
	}

	for _, v := range tests {
		p := Local{
			private: private{
				keys: [][]byte{v.key},
			},
		}

		secret := p.private.Key()

		if !reflect.DeepEqual(secret, v.expected) {
			t.Errorf("copyKey returned incorrect key\n   expected:%x\n   got:     %x", v.expected, secret)
		}
	}
}
