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
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Door struct {
	OID  catalog.OID `json:"OID"`
	Name string      `json:"name"`

	delay   uint8
	mode    core.ControlState
	created time.Time
	deleted *types.DateTime
}

const BLANK = "'blank'"

var created = time.Now()

func (d *Door) IsValid() bool {
	if d != nil {
		door := d.Door()

		if strings.TrimSpace(d.Name) != "" || door != 0 {
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
	name := d.Name

	delay := struct {
		delay      uint8
		configured uint8
		status     types.Status
		err        string
	}{
		configured: d.delay,
		status:     types.StatusUnknown,
	}

	control := struct {
		control    core.ControlState
		configured core.ControlState
		status     types.Status
		err        string
	}{
		configured: d.mode,
		status:     types.StatusUnknown,
	}

	if v := catalog.GetV(d.OID, DoorDelay); v != nil {
		delay.delay = v.(uint8)
		modified := false

		if v := catalog.GetV(d.OID, DoorDelayModified); v != nil {
			if b, ok := v.(bool); ok {
				modified = b
			}
		}

		switch {
		case modified:
			delay.status = types.StatusUncertain

		case v == d.delay:
			delay.status = types.StatusOk

		default:
			delay.status = types.StatusError
			delay.err = fmt.Sprintf("Door delay (%vs) does not match configuration (%vs)", v, d.delay)
		}
	}

	if v := catalog.GetV(d.OID, DoorControl); v != nil {
		control.control = v.(core.ControlState)
		modified := false

		if v := catalog.GetV(d.OID, DoorControlModified); v != nil {
			if b, ok := v.(bool); ok {
				modified = b
			}
		}

		switch {
		case modified:
			control.status = types.StatusUncertain

		case v == d.mode:
			control.status = types.StatusOk

		default:
			control.status = types.StatusError
			control.err = fmt.Sprintf("Door control state ('%v') does not match configuration ('%v')", v, d.mode)
		}
	}

	objects := []interface{}{
		catalog.NewObject(d.OID, ""),
		catalog.NewObject2(d.OID, DoorStatus, d.status()),
		catalog.NewObject2(d.OID, DoorCreated, created),
		catalog.NewObject2(d.OID, DoorDeleted, d.deleted),
		catalog.NewObject2(d.OID, DoorName, name),
		catalog.NewObject2(d.OID, DoorDelay, types.Uint8(delay.delay)),
		catalog.NewObject2(d.OID, DoorDelayStatus, delay.status),
		catalog.NewObject2(d.OID, DoorDelayConfigured, delay.configured),
		catalog.NewObject2(d.OID, DoorDelayError, delay.err),
		catalog.NewObject2(d.OID, DoorControl, control.control),
		catalog.NewObject2(d.OID, DoorControlStatus, control.status),
		catalog.NewObject2(d.OID, DoorControlConfigured, control.configured),
		catalog.NewObject2(d.OID, DoorControlError, control.err),
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

func (d *Door) status() types.Status {
	if d.deleted != nil {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (d *Door) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	objects := []catalog.Object{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateDoor(d, field, value)
	}

	if d != nil {
		name := fmt.Sprintf("%v", d.Name)

		switch oid {
		case d.OID.Append(DoorName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				d.log(auth, "update", d.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(d.Name, BLANK), stringify(value, BLANK)), dbc)
				d.Name = value
				objects = append(objects, catalog.NewObject2(d.OID, DoorName, d.Name))
			}

		case d.OID.Append(DoorDelay):
			delay := d.delay

			if err := f("delay", value); err != nil {
				return nil, err
			} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
				return nil, err
			} else {
				d.delay = uint8(v)
				objects = append(objects, catalog.NewObject2(d.OID, DoorDelayStatus, types.StatusUncertain))
				objects = append(objects, catalog.NewObject2(d.OID, DoorDelayConfigured, d.delay))
				objects = append(objects, catalog.NewObject2(d.OID, DoorDelayError, ""))
				objects = append(objects, catalog.NewObject2(d.OID, DoorDelayModified, true))

				d.log(auth, "update", d.OID, "delay", fmt.Sprintf("Updated delay from %vs to %vs", delay, value), dbc)
			}

		case d.OID.Append(DoorControl):
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

				objects = append(objects, catalog.NewObject2(d.OID, DoorControlStatus, types.StatusUncertain))
				objects = append(objects, catalog.NewObject2(d.OID, DoorControlConfigured, d.mode))
				objects = append(objects, catalog.NewObject2(d.OID, DoorControlError, ""))
				objects = append(objects, catalog.NewObject2(d.OID, DoorControlModified, true))

				d.log(auth, "update", d.OID, "mode", fmt.Sprintf("Updated mode from %v to %v", mode, value), dbc)
			}
		}

		if !d.IsValid() {
			if auth != nil {
				if err := auth.CanDeleteDoor(d); err != nil {
					return nil, err
				}
			}

			d.log(auth, "delete", d.OID, "name", fmt.Sprintf("Deleted door %v", name), dbc)
			now := types.DateTime(time.Now())
			d.deleted = &now

			catalog.Delete(d.OID)
		}

		objects = append(objects, catalog.NewObject2(d.OID, DoorStatus, d.status()))
		objects = append(objects, catalog.NewObject2(d.OID, DoorDeleted, d.deleted))
	}

	return objects, nil
}

func (d Door) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID       `json:"OID"`
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
		OID     catalog.OID       `json:"OID"`
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

func (d *Door) lookup(suffix catalog.Suffix) interface{} {
	return catalog.GetV(d.OID, suffix)
}

func (d *Door) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "door",
		Operation: operation,
		Details: audit.Details{
			ID:          fmt.Sprintf("%v:%v", d.DeviceID(), d.Door()),
			Name:        stringify(d.Name, ""),
			Field:       field,
			Description: description,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}

func (d *Door) DeviceID() uint32 {
	if d != nil {
		if deviceID := d.lookup(DoorControllerID); deviceID != nil {
			if v, ok := deviceID.(*uint32); ok && v != nil {
				return *v
			}
		}

	}

	return 0
}

func (d *Door) Door() uint8 {
	if d != nil {
		if door := d.lookup(DoorControllerDoor); door != nil {
			if v, ok := door.(uint8); ok {
				return v
			}
		}

	}

	return 0
}

func stringify(i interface{}, defval string) string {
	s := ""

	switch v := i.(type) {
	case *uint32:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	default:
		if i != nil {
			s = fmt.Sprintf("%v", i)
		}
	}

	if s != "" {
		return s
	}

	return defval
}
