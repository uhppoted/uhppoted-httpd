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
	OID  string `json:"OID"`
	Name string `json:"name"`

	delay   uint8
	mode    core.ControlState
	created time.Time
	deleted *time.Time
}

var created = time.Now()

func (d *Door) IsValid() bool {
	if d != nil {
		controller := ""
		if v, _ := catalog.GetV(d.OID + ".0.2"); v != nil {
			controller = stringify(v)
		}

		door := ""
		if v, _ := catalog.GetV(d.OID + ".0.2.4"); v != nil {
			door = stringify(v)
		}

		if strings.TrimSpace(d.Name) != "" || (controller != "" && door != "") {
			return true
		}
	}

	return false
}

func (d *Door) IsDeleted() bool {
	if d != nil && d.deleted != nil {
		return true
	}

	return false
}

func (d Door) String() string {
	return fmt.Sprintf("%v", d.Name)
}

func (d *Door) AsObjects() []interface{} {
	created := d.created.Format("2006-01-02 15:04:05")
	status := stringify(types.StatusOk)
	name := stringify(d.Name)

	controller := struct {
		OID     string
		created string
		name    string
		ID      string
		door    string
	}{
		OID:     stringify(d.lookup(".0.2")),
		created: stringify(d.lookup(".0.2.1")),
		name:    stringify(d.lookup(".0.2.2")),
		ID:      stringify(d.lookup(".0.2.3")),
		door:    stringify(d.lookup(".0.2.4")),
	}

	delay := struct {
		delay      string
		configured string
		status     string
		err        string
	}{
		configured: stringify(d.delay),
		status:     stringify(types.StatusUnknown),
	}

	control := struct {
		control    string
		configured string
		status     string
		err        string
	}{
		configured: stringify(d.mode),
		status:     stringify(types.StatusUnknown),
	}

	if v, dirty := catalog.GetV(d.OID + ".2"); v != nil {
		delay.delay = stringify(v)

		switch {
		case dirty:
			delay.status = stringify(types.StatusUncertain)

		case v == d.delay:
			delay.status = stringify(types.StatusOk)

		default:
			delay.status = stringify(types.StatusError)
			delay.err = fmt.Sprintf("Door delay (%vs) does not match configuration (%vs)", v, d.delay)
		}
	}

	if v, dirty := catalog.GetV(d.OID + ".3"); v != nil {
		control.control = stringify(v.(core.ControlState))

		switch {
		case dirty:
			control.status = stringify(types.StatusUncertain)

		case v == d.mode:
			control.status = stringify(types.StatusOk)

		default:
			control.status = stringify(types.StatusError)
			control.err = fmt.Sprintf("Door control state ('%v') does not match configuration ('%v')", v, d.mode)
		}
	}

	if d.deleted != nil {
		status = stringify(types.StatusDeleted)
	}

	objects := []interface{}{
		object{OID: d.OID, Value: status},
		object{OID: d.OID + ".0.1", Value: created},
		object{OID: d.OID + ".0.2", Value: controller.OID},
		object{OID: d.OID + ".0.2.1", Value: controller.created},
		object{OID: d.OID + ".0.2.2", Value: controller.name},
		object{OID: d.OID + ".0.2.3", Value: controller.ID},
		object{OID: d.OID + ".0.2.4", Value: controller.door},
		object{OID: d.OID + ".1", Value: name},
		object{OID: d.OID + ".2", Value: delay.delay},
		object{OID: d.OID + ".2.1", Value: delay.status},
		object{OID: d.OID + ".2.2", Value: delay.configured},
		object{OID: d.OID + ".2.3", Value: delay.err},
		object{OID: d.OID + ".3", Value: control.control},
		object{OID: d.OID + ".3.1", Value: control.status},
		object{OID: d.OID + ".3.2", Value: control.configured},
		object{OID: d.OID + ".3.3", Value: control.err},
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

func (d Door) serialize() ([]byte, error) {
	record := struct {
		OID     string            `json:"OID"`
		Name    string            `json:"name,omitempty"`
		Delay   uint8             `json:"delay,omitempty"`
		Mode    core.ControlState `json:"mode,omitempty"`
		Created string            `json:"created"`
	}{
		OID:     d.OID,
		Name:    d.Name,
		Delay:   d.delay,
		Mode:    d.mode,
		Created: d.created.Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(record)
}

func (d *Door) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     string            `json:"OID"`
		Name    string            `json:"name,omitempty"`
		Delay   uint8             `json:"delay,omitempty"`
		Mode    core.ControlState `json:"mode,omitempty"`
		Created string            `json:"created"`
	}{
		Delay: 5,
		Mode:  core.Controlled,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	d.OID = record.OID
	d.Name = record.Name
	d.delay = record.Delay
	d.mode = record.Mode
	d.created = created

	if t, err := time.Parse("2006-01-02 15:04:05", record.Created); err == nil {
		d.created = t
	}

	return nil
}

func (d *Door) clone() Door {
	return Door{
		OID:     d.OID,
		Name:    d.Name,
		delay:   d.delay,
		mode:    d.mode,
		created: d.created,
		deleted: d.deleted,
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
			delay := d.delay

			if err := f("delay", value); err != nil {
				return nil, err
			} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
				return nil, err
			} else {
				d.delay = uint8(v)

				catalog.PutV(d.OID+".2.2", d.delay, false)

				objects = append(objects, object{
					OID:   d.OID + ".2",
					Value: stringify(d.delay),
				})

				objects = append(objects, object{
					OID:   d.OID + ".2.1",
					Value: stringify(types.StatusUncertain),
				})

				objects = append(objects, object{
					OID:   d.OID + ".2.2",
					Value: stringify(d.delay),
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
				mode := d.mode
				switch value {
				case "controlled":
					d.mode = core.Controlled
				case "normally open":
					d.mode = core.NormallyOpen
				case "normally closed":
					d.mode = core.NormallyClosed
				default:
					return nil, fmt.Errorf("%v: invalid control state (%v)", d.Name, value)
				}

				catalog.PutV(d.OID+".3.2", d.mode, false)

				objects = append(objects, object{
					OID:   d.OID + ".3",
					Value: stringify(d.mode),
				})

				objects = append(objects, object{
					OID:   d.OID + ".3.1",
					Value: stringify(types.StatusUncertain),
				})

				objects = append(objects, object{
					OID:   d.OID + ".3.2",
					Value: stringify(d.mode),
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

func (d *Door) lookup(suffix string) interface{} {
	v, _ := catalog.GetV(d.OID + suffix)

	if v != nil {
		return v
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
		if i != nil {
			return fmt.Sprintf("%v", i)
		}
	}

	return ""
}
