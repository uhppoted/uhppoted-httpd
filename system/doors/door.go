package doors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Door struct {
	OID   string            `json:"OID"`
	Name  string            `json:"name"`
	Delay uint8             `json:"delay"`
	Mode  core.ControlState `json:"mode"`

	created time.Time
	deleted *time.Time
}

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
		delay      uint8
		configured uint8
		status     types.Status
		err        string
	}{
		configured: d.Delay,
		status:     types.StatusUnknown,
	}

	minfo := struct {
		mode       core.ControlState
		configured core.ControlState
		status     types.Status
		err        string
	}{
		configured: d.Mode,
		status:     types.StatusUnknown,
	}

	if v, ok := d.lookup("controller.Door.Delay for door.OID[%[1]v]").(uint8); ok {
		delay.delay = v
		if v == d.Delay {
			delay.status = types.StatusOk
		} else {
			delay.status = types.StatusError
			delay.err = fmt.Sprintf("Door delay (%vs) does not match configuration (%vs)", v, d.Delay)
		}
	}

	if v, ok := d.lookup("controller.Door.Mode for door.OID[%[1]v]").(core.ControlState); ok {
		minfo.mode = v
		if v == d.Mode {
			minfo.status = types.StatusOk
		} else {
			minfo.status = types.StatusError
			minfo.err = fmt.Sprintf("Door control state ('%v') does not match configuration ('%v')", v, d.Mode)
		}
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
		object{OID: d.OID + ".2", Value: stringify(delay.delay)},
		object{OID: d.OID + ".2.1", Value: stringify(delay.status)},
		object{OID: d.OID + ".2.2", Value: stringify(delay.configured)},
		object{OID: d.OID + ".2.3", Value: stringify(delay.err)},
		object{OID: d.OID + ".3", Value: stringify(minfo.mode)},
		object{OID: d.OID + ".3.1", Value: stringify(minfo.status)},
		object{OID: d.OID + ".3.2", Value: stringify(minfo.configured)},
		object{OID: d.OID + ".3.3", Value: stringify(minfo.err)},
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

func (d *Door) Get(field string) interface{} {
	if d != nil {
		switch field {
		case "Delay":
			if v := d.lookup("controller.Door.Delay for door.OID[%[1]v]"); v != nil {
				if vv, ok := v.(uint8); ok {
					return vv
				}
			}

		case "Delay.Configured":
			return d.Delay

		case "Mode":
			if v := d.lookup("controller.Door.Mode for door.OID[%[1]v]"); v != nil {
				if vv, ok := v.(core.ControlState); ok {
					return vv
				}
			}

		case "Mode.Configured":
			return d.Mode
		}
	}

	return nil
}

func (d *Door) UnmarshalJSON(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     string            `json:"OID"`
		Name    string            `json:"name,omitempty"`
		Delay   uint8             `json:"delay,omitempty"`
		Mode    core.ControlState `json:"mode,omitempty"`
		Created time.Time         `json:"created"`
	}{
		Delay:   5,
		Mode:    core.Controlled,
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

		case d.OID + ".2":
			delay := d.Delay

			if err := f("delay", value); err != nil {
				return nil, err
			} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
				return nil, err
			} else {
				d.Delay = uint8(v)

				objects = append(objects, object{
					OID:   d.OID + ".2",
					Value: stringify(d.Delay),
				})

				objects = append(objects, object{
					OID:   d.OID + ".2.1",
					Value: stringify(types.StatusUncertain),
				})

				objects = append(objects, object{
					OID:   d.OID + ".2.2",
					Value: stringify(d.Delay),
				})

				objects = append(objects, object{
					OID:   d.OID + ".2.3",
					Value: "",
				})

				d.log(auth, "update", d.OID, "delay", stringify(delay), value)
			}

		case d.OID + ".3":
			if err := f("mode", value); err != nil {
				return nil, err
			} else {
				mode := d.Mode
				switch value {
				case "controlled":
					d.Mode = core.Controlled
				case "normally open":
					d.Mode = core.NormallyOpen
				case "normally closed":
					d.Mode = core.NormallyClosed
				default:
					return nil, fmt.Errorf("%v: invalid control state (%v)", d.Name, value)
				}

				objects = append(objects, object{
					OID:   d.OID + ".3",
					Value: stringify(d.Mode),
				})

				objects = append(objects, object{
					OID:   d.OID + ".3.1",
					Value: stringify(types.StatusUncertain),
				})

				objects = append(objects, object{
					OID:   d.OID + ".3.2",
					Value: stringify(d.Mode),
				})

				objects = append(objects, object{
					OID:   d.OID + ".3.3",
					Value: "",
				})

				d.log(auth, "update", d.OID, "mode", stringify(mode), value)
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

func (d *Door) lookup(query string) interface{} {
	q := fmt.Sprintf(query, d.OID)
	v := catalog.Get(q)

	if v != nil && len(v) > 0 {
		return v[0]
	}

	return nil
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
