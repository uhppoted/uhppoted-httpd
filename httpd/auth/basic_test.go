package auth

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
)

func TestBasicDeserialize(t *testing.T) {
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

	expected := Basic{
		resources: []resource{
			resource{
				Path:       regexp.MustCompile("^[a-z]+[0-9]+$"),
				Authorised: regexp.MustCompile("[0-9]+[a-z]*"),
			},
		},
	}

	basic := Basic{}
	r := bytes.NewBuffer([]byte(json))

	if err := basic.deserialize(r); err != nil {
		t.Fatalf("Unexpected error unmarshalling auth.Basic (%v)", err)
	}

	// ... check this way because reflect.DeepEqual copies guard value
	if !reflect.DeepEqual(basic.resources, expected.resources) {
		t.Errorf("Incorrectly deserialized:\n   expected:%#v\n   got:     %#v", &expected, &basic)
	}
}
