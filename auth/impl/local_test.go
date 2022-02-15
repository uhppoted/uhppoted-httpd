package local

import (
	"bytes"
	"reflect"
	"testing"
)

func TestLocalDeserialize(t *testing.T) {
	json := `{
  "users": {
    "hagrid": {
      "salt": "efcd3412",
      "password": "dragon",
      "role": "grounds keeper"
    }
  }
}`

	expected := Local{
		keys: [][]byte{[]byte("qwerty")},
		users: map[string]*user{
			"hagrid": &user{
				Salt:     salt([]byte{0xef, 0xcd, 0x34, 0x12}),
				Password: "dragon",
				Role:     "grounds keeper",
			},
		},
	}

	local := Local{
		keys: [][]byte{[]byte("qwerty")},
	}

	r := bytes.NewBuffer([]byte(json))

	if err := local.deserialize(r); err != nil {
		t.Fatalf("Unexpected error unmarshalling auth.Local (%v)", err)
	}

	// ... check this way because reflect.DeepEqual copies guard value
	if !reflect.DeepEqual(local.keys, expected.keys) {
		t.Errorf("Incorrectly deserialized:\n   expected:%x\n   got:     %x", expected.keys, local.keys)
	}

	if !reflect.DeepEqual(local.users, expected.users) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &local)
	}
}
