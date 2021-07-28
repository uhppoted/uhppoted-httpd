package doors

import (
	"fmt"
	"time"
)

type Door struct {
	OID  string `json:"OID"`
	Name string `json:"name"`

	created time.Time
}

func (d *Door) IsValid() bool {
	return true
}

func (d *Door) AsObjects() []interface{} {
	created := d.created.Format("2006-01-02 15:04:05")
	status := StatusOk
	name := stringify(d.Name)

	objects := []interface{}{
		object{OID: d.OID, Value: fmt.Sprintf("%v", status)},
		object{OID: d.OID + ".0.1", Value: created},
		object{OID: d.OID + ".1", Value: name},
	}

	return objects
}

func (d *Door) Clone() Door {
	return Door{
		OID:  d.OID,
		Name: d.Name,
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

	default:
		return fmt.Sprintf("%v", i)
	}

	return ""
}
