package catalog

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type Object struct {
	OID   OID    `json:"OID"`
	Value string `json:"value"`
}

func NewObject(oid OID, value interface{}) Object {
	return Object{
		OID:   oid,
		Value: stringify(value),
	}
}

func NewObject2(oid OID, suffix Suffix, value interface{}) Object {
	return Object{
		OID:   oid.Append(suffix),
		Value: stringify(value),
	}
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
