package doors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Door struct {
	catalog.CatalogDoor
	name string

	delay uint8
	mode  core.ControlState

	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

var created = types.TimestampNow()

func (d Door) IsValid() bool {
	return d.validate() == nil
}

func (d Door) validate() error {
	door := catalog.GetDoorDeviceDoor(d.OID)

	if strings.TrimSpace(d.name) == "" && door == 0 {
		return fmt.Errorf("Door name cannot be blank unless door is assigned to a controller")
	}

	return nil
}

func (d *Door) IsDeleted() bool {
	return !d.deleted.IsZero()
}

func (d Door) IsOk() bool {
	mode := types.StatusUnknown
	delay := types.StatusUnknown

	if v := catalog.GetV(d.OID, DoorControl); v != nil {
		if b, ok := catalog.GetBool(d.OID, DoorControlModified); ok && b {
			mode = types.StatusUncertain
		} else if d.mode == v.(core.ControlState) {
			mode = types.StatusOk
		} else {
			mode = types.StatusError
		}
	}

	if v, ok := catalog.GetUint8(d.OID, DoorDelay); ok {
		if b, ok := catalog.GetBool(d.OID, DoorDelayModified); ok && b {
			delay = types.StatusUncertain
		} else if v == d.delay {
			delay = types.StatusOk
		} else {
			delay = types.StatusError
		}
	}

	return mode != types.StatusError && delay != types.StatusError
}

func (d *Door) Mode() core.ControlState {
	if d != nil {
		return d.mode
	}

	return core.ModeUnknown
}

func (d *Door) Delay() uint8 {
	if d != nil {
		return d.delay
	}

	return 0
}

func (d Door) String() string {
	return d.name
}

func (d *Door) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if d.IsDeleted() {
		list = append(list, kv{DoorDeleted, d.deleted})
	} else {
		name := d.name

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

		if v, ok := catalog.GetUint8(d.OID, DoorDelay); ok {
			delay.delay = v
			modified := false

			if b, ok := catalog.GetBool(d.OID, DoorDelayModified); ok {
				modified = b
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

			if b, ok := catalog.GetBool(d.OID, DoorControlModified); ok {
				modified = b
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

		list = append(list, kv{DoorStatus, d.Status()})
		list = append(list, kv{DoorCreated, d.created})
		list = append(list, kv{DoorDeleted, d.deleted})
		list = append(list, kv{DoorName, name})
		list = append(list, kv{DoorDelay, types.Uint8(delay.delay)})
		list = append(list, kv{DoorDelayStatus, delay.status})
		list = append(list, kv{DoorDelayConfigured, delay.configured})
		list = append(list, kv{DoorDelayError, delay.err})
		list = append(list, kv{DoorControl, control.control})
		list = append(list, kv{DoorControlStatus, control.status})
		list = append(list, kv{DoorControlConfigured, control.configured})
		list = append(list, kv{DoorControlError, control.err})
	}

	return d.toObjects(list, a)
}

func (d Door) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Name string
	}{
		Name: d.name,
	}

	return "door", &entity
}

func (d *Door) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	f := func(field string, value interface{}) error {
		if a == nil {
			return nil
		}

		return a.CanUpdate(d, field, value, auth.Doors)
	}

	if d == nil {
		return []schema.Object{}, nil
	} else if d.IsDeleted() {
		return d.toObjects([]kv{kv{DoorDeleted, d.deleted}}, a), fmt.Errorf("Door has been deleted")
	}

	uid := auth.UID(a)
	list := []kv{}

	switch oid {
	case d.OID.Append(DoorName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			d.log(dbc, uid, "update", "name", d.name, value, "Updated name from %v to %v", d.name, value)

			d.name = value
			d.modified = types.TimestampNow()

			list = append(list, kv{DoorName, d.name})
		}

	case d.OID.Append(DoorDelay):
		delay := d.delay

		if err := f("delay", value); err != nil {
			return nil, err
		} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
			return nil, err
		} else {
			d.delay = uint8(v)
			d.modified = types.TimestampNow()

			list = append(list, kv{DoorDelayStatus, types.StatusUncertain})
			list = append(list, kv{DoorDelayConfigured, d.delay})
			list = append(list, kv{DoorDelayError, ""})
			list = append(list, kv{DoorDelayModified, true})

			dbc.Updated(d.OID, DoorDelay, d.delay)

			d.log(dbc, uid, "update", "delay", delay, value, "Updated delay from %vs to %vs", delay, value)
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
				return nil, fmt.Errorf("%v: invalid control state (%v)", d.name, value)
			}

			d.modified = types.TimestampNow()

			list = append(list, kv{DoorControlStatus, types.StatusUncertain})
			list = append(list, kv{DoorControlConfigured, d.mode})
			list = append(list, kv{DoorControlError, ""})
			list = append(list, kv{DoorControlModified, true})

			dbc.Updated(d.OID, DoorControl, d.mode)

			d.log(dbc, uid, "update", "mode", mode, value, "Updated mode from %v to %v", mode, value)
		}
	}

	list = append(list, kv{DoorStatus, d.Status()})

	return d.toObjects(list, a), nil
}

func (d *Door) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if d != nil {
		if a != nil {
			if err := a.CanDelete(d, auth.Doors); err != nil {
				return nil, err
			}
		}

		if door := catalog.GetDoorDeviceDoor(d.OID); door != 0 {
			return nil, fmt.Errorf("Cannot delete door %v - assigned to controller", d.name)
		}

		d.log(dbc, auth.UID(a), "delete", "name", d.name, "", "Deleted door %v", d.name)
		d.deleted = types.TimestampNow()
		d.modified = types.TimestampNow()

		list = append(list, kv{DoorDeleted, d.deleted})
		list = append(list, kv{DoorStatus, d.Status()})

		catalog.DeleteT(d.CatalogDoor, d.OID)
	}

	return d.toObjects(list, a), nil
}

func (d *Door) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	f := func(d *Door, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(d, field, value, auth.Doors); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if !d.IsDeleted() && f(d, "OID", d.OID) {
		objects = append(objects, catalog.NewObject(d.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(d, field, v.value) {
			objects = append(objects, catalog.NewObject2(d.OID, v.field, v.value))
		}
	}

	return objects
}

func (d Door) Status() types.Status {
	if d.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (d Door) serialize() ([]byte, error) {
	record := struct {
		OID      schema.OID        `json:"OID"`
		Name     string            `json:"name,omitempty"`
		Delay    uint8             `json:"delay,omitempty"`
		Mode     core.ControlState `json:"mode,omitempty"`
		Created  types.Timestamp   `json:"created,omitempty"`
		Modified types.Timestamp   `json:"modified,omitempty"`
	}{
		OID:      d.OID,
		Name:     d.name,
		Delay:    d.delay,
		Mode:     d.mode,
		Created:  d.created.UTC(),
		Modified: d.modified.UTC(),
	}

	return json.Marshal(record)
}

func (d *Door) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      schema.OID        `json:"OID"`
		Name     string            `json:"name,omitempty"`
		Delay    uint8             `json:"delay,omitempty"`
		Mode     core.ControlState `json:"mode,omitempty"`
		Created  types.Timestamp   `json:"created,omitempty"`
		Modified types.Timestamp   `json:"modified,omitempty"`
	}{
		Delay:   5,
		Mode:    core.Controlled,
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	d.OID = record.OID
	d.name = record.Name
	d.delay = record.Delay
	d.mode = record.Mode
	d.created = record.Created
	d.modified = record.Modified

	return nil
}

func (d *Door) clone() Door {
	return Door{
		CatalogDoor: catalog.CatalogDoor{
			OID: d.OID,
		},
		name:     d.name,
		delay:    d.delay,
		mode:     d.mode,
		created:  d.created,
		modified: d.modified,
		deleted:  d.deleted,
	}
}

func (d *Door) log(dbc db.DBC, uid string, operation string, field string, before, after any, format string, fields ...any) {
	deviceID := catalog.GetDoorDeviceID(d.OID)
	door := catalog.GetDoorDeviceDoor(d.OID)
	ID := fmt.Sprintf("%v/%v", deviceID, door)
	name := d.name

	dbc.Log(uid, operation, d.OID, "door", ID, name, field, before, after, format, fields...)
}
