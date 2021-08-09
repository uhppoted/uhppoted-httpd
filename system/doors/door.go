package doors

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Door struct {
	OID   string `json:"OID"`
	Name  string `json:"name"`
	Delay uint8  `json:"delay"`
	Mode  mode   `json:"mode"`

	created time.Time
	deleted *time.Time
}

type mode string

var created = time.Now()

func (d *Door) IsValid() bool {
	if d != nil {
		controller := d.lookup("controller.OID for door.OID[%v]")
		door := d.lookup("controller.Door for door.OID[%v]")

		if strings.TrimSpace(d.Name) != "" || (controller != "" && door != "") {
			return true
		}
	}

	return false
}

func (d *Door) AsObjects() []interface{} {
	delay := struct {
		delay      string
		configured uint8
		status     types.Status
	}{
		delay:      d.lookup("controller.Door.Delay for door.OID[%[1]v]"),
		configured: d.Delay,
		status:     types.StatusUnknown,
	}

	mode := struct {
		mode       string
		configured string
		status     types.Status
	}{
		mode:       d.lookup("controller.Door.Mode for door.OID[%[1]v]"),
		configured: stringify(d.Mode),
		status:     types.StatusUnknown,
	}

	created := d.created.Format("2006-01-02 15:04:05")
	status := types.StatusOk
	name := stringify(d.Name)

	controllerOID := d.lookup("controller.OID for door.OID[%[1]v]")
	controllerCreated := d.lookup("controller.Created for door.OID[%[1]v]")
	controllerName := d.lookup("controller.Name for door.OID[%[1]v]")
	controllerID := d.lookup("controller.ID for door.OID[%[1]v]")
	controllerDoor := d.lookup("controller.Door for door.OID[%[1]v]")

	objects := []interface{}{
		object{OID: d.OID, Value: fmt.Sprintf("%v", status)},
		object{OID: d.OID + ".0.1", Value: created},
		object{OID: d.OID + ".0.2", Value: stringify(controllerOID)},
		object{OID: d.OID + ".0.2.1", Value: stringify(controllerCreated)},
		object{OID: d.OID + ".0.2.2", Value: stringify(controllerName)},
		object{OID: d.OID + ".0.2.3", Value: stringify(controllerID)},
		object{OID: d.OID + ".0.2.4", Value: stringify(controllerDoor)},
		object{OID: d.OID + ".1", Value: name},
		object{OID: d.OID + ".2", Value: delay.delay},
		object{OID: d.OID + ".2.1", Value: stringify(delay.status)},
		object{OID: d.OID + ".2.2", Value: stringify(delay.configured)},
		object{OID: d.OID + ".3", Value: mode.mode},
		object{OID: d.OID + ".3.1", Value: stringify(mode.status)},
		object{OID: d.OID + ".3.2", Value: stringify(mode.configured)},
	}

	return objects
}

func (d *Door) AsRuleEntity() interface{} {
	type entity struct {
		Name string
	}

	if d != nil {
		return &entity{
			Name: fmt.Sprintf("%v", d.Name),
		}
	}

	return &entity{}
}

func (d *Door) UnmarshalJSON(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     string    `json:"OID"`
		Name    string    `json:"name,omitempty"`
		Delay   uint8     `json:"delay,omitempty"`
		Mode    mode      `json:"mode,omitempty"`
		Created time.Time `json:"created"`
	}{
		Delay:   5,
		Mode:    "controlled",
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	d.OID = record.OID
	d.Name = record.Name
	d.Delay = record.Delay
	d.Mode = record.Mode
	d.created = record.Created

	return nil
}

func (d *Door) clone() Door {
	return Door{
		OID:     d.OID,
		Name:    d.Name,
		Delay:   d.Delay,
		Mode:    d.Mode,
		created: d.created,
	}
}

func (d *Door) set(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateDoor(d, field, value)
	}

	if d != nil {
		name := stringify(d.Name)

		switch oid {
		case d.OID + ".1":
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				d.log(auth, "update", d.OID, "name", stringify(d.Name), value)
				d.Name = value
				objects = append(objects, object{
					OID:   d.OID + ".1",
					Value: stringify(d.Name),
				})
			}
		}

		if !d.IsValid() {
			if auth != nil {
				if err := auth.CanDeleteDoor(d); err != nil {
					return nil, err
				}
			}

			d.log(auth, "delete", d.OID, "name", name, "")
			now := time.Now()
			d.deleted = &now

			objects = append(objects, object{
				OID:   d.OID,
				Value: "deleted",
			})

			catalog.Delete(d.OID)
		}
	}

	return objects, nil
}

func (d *Door) lookup(query string) string {
	q := fmt.Sprintf(query, d.OID)
	v := catalog.Get(q)

	if v != nil && len(v) > 0 {
		return stringify(v[0])
	}

	return ""
}

func (d *Door) log(auth auth.OpAuth, operation, OID, field, current, value string) {
	type info struct {
		OID     string `json:"OID"`
		Door    string `json:"door"`
		Field   string `json:"field"`
		Current string `json:"current"`
		Updated string `json:"new"`
	}

	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	if trail != nil {
		record := audit.LogEntry{
			UID:       uid,
			Module:    OID,
			Operation: operation,
			Info: info{
				OID:     OID,
				Door:    stringify(d.Name),
				Field:   field,
				Current: current,
				Updated: value,
			},
		}

		trail.Write(record)
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

func (m *mode) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	switch s {
	case "controlled":
		*m = mode("controlled")
	case "normally open":
		*m = mode("normally open")
	case "normally closed":
		*m = mode("normally closed")

	default:
		return fmt.Errorf("Invalid door control state ('%v')", s)
	}

	return nil
}
