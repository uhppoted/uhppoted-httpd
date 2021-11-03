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
	OID   catalog.OID `json:"OID"`
	Name  string      `json:"name"`
	Index uint32      `json:"index"`

	delay   uint8
	mode    core.ControlState
	created time.Time
	deleted *time.Time
}

var created = time.Now()

const DoorCreated = catalog.DoorCreated
const DoorControllerOID = catalog.DoorControllerOID
const DoorControllerCreated = catalog.DoorControllerCreated
const DoorControllerName = catalog.DoorControllerName
const DoorControllerID = catalog.DoorControllerID
const DoorControllerDoor = catalog.DoorControllerDoor
const DoorName = catalog.DoorName
const DoorDelay = catalog.DoorDelay
const DoorDelayStatus = catalog.DoorDelayStatus
const DoorDelayConfigured = catalog.DoorDelayConfigured
const DoorDelayError = catalog.DoorDelayError
const DoorDelayModified = catalog.DoorDelayModified
const DoorControl = catalog.DoorControl
const DoorControlStatus = catalog.DoorControlStatus
const DoorControlConfigured = catalog.DoorControlConfigured
const DoorControlError = catalog.DoorControlError
const DoorControlModified = catalog.DoorControlModified
const DoorIndex = catalog.DoorIndex

func (d *Door) IsValid() bool {
	if d != nil {
		controller := ""
		if v := catalog.GetV(d.OID.Append(DoorControllerOID)); v != nil {
			controller = fmt.Sprintf("%v", v)
		}

		door := uint8(0)
		if v := catalog.GetV(d.OID.Append(DoorControllerDoor)); v != nil {
			door = v.(uint8)
		}

		if strings.TrimSpace(d.Name) != "" || (controller != "" && door >= 1 && door <= 4) {
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
func (d Door) String() string {
	return fmt.Sprintf("%v", d.Name)
}

func (d *Door) AsObjects() []interface{} {
	created := d.created.Format("2006-01-02 15:04:05")
	status := types.StatusOk
	name := d.Name
	index := d.Index

	controller := struct {
		OID     string
		created string
		name    string
		ID      string
		door    string
	}{
		OID:     stringify(d.lookup(DoorControllerOID), ""),
		created: stringify(d.lookup(DoorControllerCreated), ""),
		name:    stringify(d.lookup(DoorControllerName), ""),
		ID:      stringify(d.lookup(DoorControllerID), ""),
		door:    stringify(d.lookup(DoorControllerDoor), ""),
	}

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

	if v := catalog.GetV(d.OID.Append(DoorDelay)); v != nil {
		delay.delay = v.(uint8)
		modified := false

		if v := catalog.GetV(d.OID.Append(DoorDelayModified)); v != nil {
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

	if v := catalog.GetV(d.OID.Append(DoorControl)); v != nil {
		control.control = v.(core.ControlState)
		modified := false

		if v := catalog.GetV(d.OID.Append(DoorControlModified)); v != nil {
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

	if d.deleted != nil {
		status = types.StatusDeleted
	}

	objects := []interface{}{
		catalog.NewObject(d.OID, status),
		catalog.NewObject2(d.OID, DoorCreated, created),
		catalog.NewObject2(d.OID, DoorControllerOID, controller.OID),
		catalog.NewObject2(d.OID, DoorControllerCreated, controller.created),
		catalog.NewObject2(d.OID, DoorControllerName, controller.name),
		catalog.NewObject2(d.OID, DoorControllerID, controller.ID),
		catalog.NewObject2(d.OID, DoorControllerDoor, controller.door),
		catalog.NewObject2(d.OID, DoorName, name),
		catalog.NewObject2(d.OID, DoorDelay, types.Uint8(delay.delay)),
		catalog.NewObject2(d.OID, DoorDelayStatus, delay.status),
		catalog.NewObject2(d.OID, DoorDelayConfigured, delay.configured),
		catalog.NewObject2(d.OID, DoorDelayError, delay.err),
		catalog.NewObject2(d.OID, DoorControl, control.control),
		catalog.NewObject2(d.OID, DoorControlStatus, control.status),
		catalog.NewObject2(d.OID, DoorControlConfigured, control.configured),
		catalog.NewObject2(d.OID, DoorControlError, control.err),
		catalog.NewObject2(d.OID, DoorIndex, index),
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
				d.log(auth, "update", d.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(d.Name, "<blank>"), stringify(value, "<blank>")), dbc)
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
				// objects = append(objects, catalog.NewObject2(d.OID, DoorDelay, d.delay)) // TODO incorrect: fix this in JS
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

				// objects = append(objects, catalog.NewObject2(d.OID, DoorControl, d.mode)) // TODO incorrect: fix this in JS
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
			now := time.Now()
			d.deleted = &now

			objects = append(objects, catalog.NewObject(d.OID, "deleted"))

			catalog.Delete(d.OID)
		}
	}

	return objects, nil
}

func (d Door) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID       `json:"OID"`
		Name    string            `json:"name,omitempty"`
		Delay   uint8             `json:"delay,omitempty"`
		Mode    core.ControlState `json:"mode,omitempty"`
		Index   uint32            `json:"index,omitempty"`
		Created string            `json:"created"`
	}{
		OID:     d.OID,
		Name:    d.Name,
		Delay:   d.delay,
		Mode:    d.mode,
		Index:   d.Index,
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
		Index   uint32            `json:"index,omitempty"`
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
	d.Index = record.Index
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
		Index:   d.Index,
		delay:   d.delay,
		mode:    d.mode,
		created: d.created,
		deleted: d.deleted,
	}
}

func (d *Door) lookup(suffix catalog.Suffix) interface{} {
	return catalog.GetV(d.OID.Append(suffix))
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
