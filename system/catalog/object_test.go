package catalog

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestObjectToJSON(t *testing.T) {
	tests := []struct {
		object Object
		json   string
	}{
		{Object{OID: "0.1.2.1", Value: "qwerty"}, `{"OID":"0.1.2.1","value":"qwerty"}`},
		{Object{OID: "0.1.2.2", Value: ""}, `{"OID":"0.1.2.2","value":""}`},
		{Object{OID: "0.1.2.3", Value: nil}, `{"OID":"0.1.2.3","value":""}`},
		{Object{OID: "0.1.2.4", Value: 654}, `{"OID":"0.1.2.4","value":"654"}`},
	}

	for _, test := range tests {
		s, err := json.Marshal(test.object)
		if err != nil {
			t.Fatalf("Unexpected error marshalling object %v (%v)", test.object, err)
		}

		if string(s) != test.json {
			t.Errorf("Object %v incorrectly marshalled:\n   expected:%s\n   got:     %s", test.object, test.json, s)
		}
	}
}

func TestJSONToObject(t *testing.T) {
	tests := []struct {
		object Object
		json   string
	}{
		{Object{OID: "0.1.2.1", Value: "qwerty"}, `{"OID":"0.1.2.1","value":"qwerty"}`},
		{Object{OID: "0.1.2.2", Value: ""}, `{"OID":"0.1.2.2","value":""}`},
		{Object{OID: "0.1.2.3", Value: ""}, `{"OID":"0.1.2.3","value":""}`},
		{Object{OID: "0.1.2.4", Value: "654"}, `{"OID":"0.1.2.4","value":"654"}`},
	}

	for _, test := range tests {
		var object Object

		err := json.Unmarshal([]byte(test.json), &object)
		if err != nil {
			t.Fatalf("Unexpected error unmarshalling JSON %v (%v)", test.json, err)
		}

		if !reflect.DeepEqual(object, test.object) {
			t.Errorf("Object JSON %v incorrectly unmarshalled:\n   expected:%s\n   got:     %s", test.json, test.object, object)
		}
	}
}
