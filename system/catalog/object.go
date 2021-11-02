package catalog

import (
	"encoding/json"
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type Object struct {
	OID   OID
	Value interface{}
}

func NewObject(oid OID, value interface{}) Object {
	return Object{
		OID:   oid,
		Value: value,
	}
}

func NewObject2(oid OID, suffix Suffix, value interface{}) Object {
	return Object{
		OID:   oid.Append(suffix),
		Value: value,
	}
}

func (o Object) MarshalJSON() ([]byte, error) {
	v := struct {
		OID   string `json:"OID"`
		Value string `json:"value"`
	}{
		OID:   stringify(o.OID),
		Value: stringify(o.Value),
	}

	return json.Marshal(v)
}

func (o *Object) UnmarshalJSON(b []byte) error {
	v := struct {
		OID   OID    `json:"OID"`
		Value string `json:"value"`
	}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	o.OID = v.OID
	o.Value = v.Value

	return nil
}

func stringify(i interface{}) string {
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case types.DateTime:
		w := i.(types.DateTime)
		return fmt.Sprintf("%v", &w)

	default:
		if i != nil {
			return fmt.Sprintf("%v", i)
		}
	}

	return ""
}
