package auth

import (
	"reflect"
	"regexp"
	"testing"
)

func TestLocalSerialize(t *testing.T) {
	l := Local{
		key: "qwerty",
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
	}

	expected := `{
  "key": "qwerty",
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
  "key": "qwerty",
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
		key: "qwerty",
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
	}

	var local Local

	if err := local.deserialize([]byte(json)); err != nil {
		t.Fatalf("Unexpected error unmarshalling auth.Local (%v)", err)
	}

	// ... check this way because reflect.DeepEqual copies guard value
	if local.key != expected.key {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}

	if !reflect.DeepEqual(local.users, expected.users) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}

	if !reflect.DeepEqual(local.resources, expected.resources) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}
}
