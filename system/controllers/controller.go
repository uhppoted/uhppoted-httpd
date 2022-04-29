package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	catalog.CatalogController
	name     string
	IP       *core.Address
	doors    map[uint8]schema.OID
	timezone string

	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp
}

type icontroller struct {
	oid      schema.OID
	name     string
	id       uint32
	endpoint *net.UDPAddr
	timezone *time.Location
	doors    map[uint8]schema.OID
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

type cached struct {
	touched  time.Time
	address  *core.Address
	datetime struct {
		datetime core.DateTime
		modified bool
	}
	cards  *uint32
	events struct {
		status  types.Status
		first   types.Uint32
		last    types.Uint32
		current types.Uint32
	}
	acl types.Status
}

var created = types.TimestampNow()

func (c Controller) IsValid() bool {
	return c.validate() == nil
}

func (c Controller) validate() error {
	if strings.TrimSpace(c.name) == "" && c.DeviceID == 0 {
		return fmt.Errorf("At least one of controller name and device ID must be valid")
	}

	return nil
}

func (c Controller) realized() bool {
	if c.DeviceID != 0 && !c.IsDeleted() {
		return true
	}

	return false
}

func (c Controller) IsDeleted() bool {
	return !c.deleted.IsZero()
}

func (c Controller) Doors() map[uint8]schema.OID {
	return c.doors
}

func (c *Controller) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if c.IsDeleted() {
		list = append(list, kv{ControllerDeleted, c.deleted})
	} else {
		type addr struct {
			address    string
			configured string
			status     types.Status
		}

		type tinfo struct {
			datetime string
			system   string
			status   types.Status
		}

		type cinfo struct {
			cards  string
			status types.Status
		}

		type einfo struct {
			first   types.Uint32
			last    types.Uint32
			current types.Uint32
			status  types.Status
		}

		name := c.name
		deviceID := ""
		address := addr{}
		datetime := tinfo{}
		cards := cinfo{}
		events := einfo{}

		doors := map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""}

		if c.DeviceID != 0 {
			deviceID = fmt.Sprintf("%v", c.DeviceID)
		}

		if c.IP != nil {
			address.address = fmt.Sprintf("%v", c.IP)
			address.configured = fmt.Sprintf("%v", c.IP)
		}

		for _, i := range []uint8{1, 2, 3, 4} {
			if d, ok := c.doors[i]; ok {
				doors[i] = d
			}
		}

		if c.DeviceID != 0 {
			if cached := c.get(); cached != nil {
				// ... get IP address field from cached value
				if cached.address != nil {
					address.address = fmt.Sprintf("%v", cached.address)
					switch {
					case c.IP == nil || (c.IP != nil && cached.address.Equal(c.IP)):
						address.status = types.StatusOk

					case c.IP != nil && !cached.address.Equal(c.IP):
						address.status = types.StatusError

					default:
						address.status = types.StatusUnknown
					}
				}

				// ... get system date/time field from cached value
				if !cached.datetime.datetime.IsZero() {
					tz, err := types.Timezone(c.timezone)
					if err != nil {
						tz = time.Local
					}

					now := time.Now().In(tz)
					t := time.Time(cached.datetime.datetime)
					T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
					delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

					datetime.datetime = T.Format("2006-01-02 15:04:05 MST")
					datetime.system = now.Format("2006-01-02 15:04:05 MST")

					switch {
					case cached.datetime.modified:
						datetime.status = types.StatusUncertain

					case delta <= math.Abs(windows.systime.Seconds()):
						datetime.status = types.StatusOk
					default:
						datetime.status = types.StatusError
					}
				}

				// ... get ACL field from cached value
				if cached.cards != nil {
					cards.cards = fmt.Sprintf("%d", *cached.cards)
					if cached.acl == types.StatusUnknown {
						cards.status = types.StatusUncertain
					} else {
						cards.status = cached.acl
					}
				}

				// ... get events field from cached value
				events.status = cached.events.status
				events.first = cached.events.first
				events.last = cached.events.last
				events.current = cached.events.current
			}
		}

		list = append(list, kv{ControllerStatus, c.Status()})
		list = append(list, kv{ControllerCreated, c.created})
		list = append(list, kv{ControllerDeleted, c.deleted})
		list = append(list, kv{ControllerName, name})
		list = append(list, kv{ControllerDeviceID, deviceID})
		list = append(list, kv{ControllerEndpointStatus, address.status})
		list = append(list, kv{ControllerEndpointAddress, address.address})
		list = append(list, kv{ControllerEndpointConfigured, address.configured})
		list = append(list, kv{ControllerDateTimeStatus, datetime.status})
		list = append(list, kv{ControllerDateTimeCurrent, datetime.datetime})
		list = append(list, kv{ControllerDateTimeSystem, datetime.system})
		list = append(list, kv{ControllerCardsStatus, cards.status})
		list = append(list, kv{ControllerCardsCount, cards.cards})
		list = append(list, kv{ControllerEventsStatus, events.status})
		list = append(list, kv{ControllerEventsFirst, events.first})
		list = append(list, kv{ControllerEventsLast, events.last})
		list = append(list, kv{ControllerEventsCurrent, events.current})
		list = append(list, kv{ControllerDoor1, doors[1]})
		list = append(list, kv{ControllerDoor2, doors[2]})
		list = append(list, kv{ControllerDoor3, doors[3]})
		list = append(list, kv{ControllerDoor4, doors[4]})
	}

	return c.toObjects(list, a)
}

func (c *Controller) AsRuleEntity() (string, interface{}) {
	v := struct {
		Name     string
		DeviceID uint32
	}{}

	if c != nil {
		v.Name = c.name
		v.DeviceID = c.DeviceID
	}

	return "controller", &v
}

func (c *Controller) AsIController() types.IController {
	var endpoint *net.UDPAddr
	var location *time.Location = time.Local
	var doors = map[uint8]schema.OID{}

	if c.IP != nil {
		endpoint = (*net.UDPAddr)(c.IP)
	}

	if tz, err := types.Timezone(c.timezone); err == nil && tz != nil {
		location = tz
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.doors[d]; ok {
			doors[d] = oid
		}
	}

	return &icontroller{
		oid:      c.OID,
		name:     c.name,
		id:       c.DeviceID,
		endpoint: endpoint,
		timezone: location,
		doors:    doors,
	}
}

func (c *Controller) String() string {
	if c == nil {
		return ""
	}

	if deviceID := c.DeviceID; deviceID == 0 {
		return fmt.Sprintf("%v", c.name)
	} else {
		return fmt.Sprintf("%v (%v)", c.name, deviceID)
	}
}

func (c Controller) Status() types.Status {
	if c.IsDeleted() {
		return types.StatusDeleted
	}

	if c.DeviceID != 0 {
		if cached := c.get(); cached != nil {
			dt := time.Now().Sub(cached.touched)
			switch {
			case dt < windows.deviceOk:
				return types.StatusOk

			case dt < windows.deviceUncertain:
				return types.StatusUncertain
			}
		}
	}

	return types.StatusUnknown
}

func (c *Controller) get() *cached {
	e := cached{
		acl: types.StatusUnknown,
	}

	e.events.status = types.StatusUnknown
	e.events.first = 0
	e.events.last = 0
	e.events.current = 0

	if v := catalog.GetV(c.OID, ControllerTouched); v != nil {
		if touched, ok := v.(time.Time); ok {
			e.touched = touched
		}
	}

	if v := catalog.GetV(c.OID, ControllerEndpointAddress); v != nil {
		if address, ok := v.(core.Address); ok {
			e.address = &address
		}
	}

	if v := catalog.GetV(c.OID, ControllerDateTimeCurrent); v != nil {
		if datetime, ok := v.(core.DateTime); ok {
			e.datetime.datetime = datetime
		}
	}

	if v := catalog.GetV(c.OID, ControllerDateTimeModified); v != nil {
		if b, ok := v.(bool); ok {
			e.datetime.modified = b
		}
	}

	if v := catalog.GetV(c.OID, ControllerCardsCount); v != nil {
		if cards, ok := v.(uint32); ok {
			e.cards = &cards
		}
	}

	if v := catalog.GetV(c.OID, ControllerEventsStatus); v != nil {
		if status, ok := v.(types.Status); ok {
			e.events.status = status
		}
	}

	if v := catalog.GetV(c.OID, ControllerEventsFirst); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.first = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.OID, ControllerEventsLast); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.last = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.OID, ControllerEventsCurrent); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.current = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.OID, ControllerCardsStatus); v != nil {
		if acl, ok := v.(types.Status); ok {
			e.acl = acl
		}
	}

	return &e
}

func (c *Controller) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if c == nil {
		return []schema.Object{}, nil
	}

	if c.IsDeleted() {
		return c.toObjects([]kv{{ControllerDeleted, c.deleted}}, a), fmt.Errorf("Controller has been deleted")
	}

	f := func(field string, value interface{}) error {
		if a != nil {
			return a.CanUpdate(c, field, value, auth.Controllers)
		}

		return nil
	}

	uid := auth.UID(a)
	OID := c.OID
	clone := c.clone()
	list := []kv{}

	switch oid {
	case OID.Append(ControllerName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			c.name = strings.TrimSpace(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerName, c.name})

			c.updated(uid, "name", clone.name, c.name, dbc)
		}

	case OID.Append(ControllerDeviceID):
		if err := f("deviceID", value); err != nil {
			return nil, err
		} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
			if id, err := strconv.ParseUint(value, 10, 32); err == nil {
				c.DeviceID = uint32(id)
				c.modified = types.TimestampNow()

				list = append(list, kv{ControllerDeviceID, c.DeviceID})
				c.updated(uid, "device-id", clone.DeviceID, c.DeviceID, dbc)
			}
		} else if value == "" {
			if p := stringify(c.DeviceID, ""); p != "" {
				c.log(uid, "update", OID, "device-id", fmt.Sprintf("Cleared device ID %v", p), p, "", dbc)
			} else if p = stringify(c.name, ""); p != "" {
				c.log(uid, "update", OID, "device-id", fmt.Sprintf("Cleared device ID for %v", p), "", "", dbc)
			} else {
				c.log(uid, "update", OID, "device-id", fmt.Sprintf("Cleared device ID"), "", "", dbc)
			}

			c.DeviceID = 0
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDeviceID, ""})
		}

	case OID.Append(ControllerEndpointAddress):
		if addr, err := core.ResolveAddr(value); err != nil {
			return nil, err
		} else if err := f("address", addr); err != nil {
			return nil, err
		} else {
			c.IP = addr
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerEndpointAddress, addr})
			list = append(list, kv{ControllerEndpointConfigured, addr})
			list = append(list, kv{ControllerEndpointStatus, types.StatusUncertain})

			c.updated(uid, "address", clone.IP, c.IP, dbc)
		}

	case OID.Append(ControllerDateTimeCurrent):
		if tz, err := types.Timezone(value); err != nil {
			return nil, err
		} else if err := f("timezone", tz); err != nil {
			return nil, err
		} else {
			c.timezone = tz.String()
			c.modified = types.TimestampNow()

			if c.DeviceID != 0 {
				if cached := c.get(); cached != nil {
					if !cached.datetime.datetime.IsZero() {
						tz, err := types.Timezone(c.timezone)
						if err != nil {
							tz = time.Local
						}

						dt := time.Now().In(tz)

						list = append(list, kv{ControllerDateTimeStatus, types.StatusUncertain})
						list = append(list, kv{ControllerDateTime, dt.Format("2006-01-02 15:04 MST")})
						list = append(list, kv{ControllerDateTimeModified, true})

						dbc.Updated(c.AsIController(), ControllerDateTime, dt)
					}
				}
			}

			c.updated(uid, "timezone", clone.timezone, c.timezone, dbc)
		}

	case OID.Append(ControllerDoor1):
		if err := f("door[1]", value); err != nil {
			return nil, err
		} else {
			c.doors[1] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor1, c.doors[1]})

			c.updated(uid, "door:1", clone.doors[1], c.doors[1], dbc)
		}

	case OID.Append(ControllerDoor2):
		if err := f("door[2]", value); err != nil {
			return nil, err
		} else {
			c.doors[2] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor2, c.doors[2]})

			c.updated(uid, "door:2", clone.doors[2], c.doors[2], dbc)
		}

	case OID.Append(ControllerDoor3):
		if err := f("door[3]", value); err != nil {
			return nil, err
		} else {
			c.doors[3] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor3, c.doors[3]})

			c.updated(uid, "door:3", clone.doors[3], c.doors[3], dbc)
		}

	case OID.Append(ControllerDoor4):
		if err := f("door[4]", value); err != nil {
			return nil, err
		} else {
			c.doors[4] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor4, c.doors[4]})

			c.updated(uid, "door:4", clone.doors[4], c.doors[4], dbc)
		}
	}

	list = append(list, kv{ControllerStatus, c.Status()})

	return c.toObjects(list, a), nil
}

func (c *Controller) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if c != nil {
		uid := auth.UID(a)

		if a != nil {
			if err := a.CanDelete(c, auth.Controllers); err != nil {
				return nil, err
			}
		}

		if p := stringify(c.name, ""); p != "" {
			c.log(uid, "delete", c.OID, "device-id", fmt.Sprintf("Deleted controller %v", p), "", "", dbc)
		} else if p = stringify(c.DeviceID, ""); p != "" {
			c.log(uid, "delete", c.OID, "device-id", fmt.Sprintf("Deleted controller %v", p), "", "", dbc)
		} else {
			c.log(uid, "delete", c.OID, "device-id", fmt.Sprintf("Deleted controller"), "", "", dbc)
		}

		c.deleted = types.TimestampNow()
		c.modified = types.TimestampNow()

		list = append(list, kv{ControllerDeleted, c.deleted})
		list = append(list, kv{ControllerStatus, c.Status()})

		catalog.DeleteT(c.CatalogController, c.OID)
	}

	return c.toObjects(list, a), nil
}

func (c *Controller) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	f := func(c *Controller, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(c, field, value, auth.Controllers); err != nil {
				return false
			}
		}

		return true
	}

	OID := c.OID
	objects := []schema.Object{}

	if !c.IsDeleted() && f(c, "OID", OID) {
		catalog.Join(&objects, catalog.NewObject(OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(c, field, v.value) {
			catalog.Join(&objects, catalog.NewObject2(OID, v.field, v.value))
		}
	}

	return objects
}

func (c *Controller) refreshed() {
	expired := time.Now().Add(-windows.cacheExpiry)

	touched := time.Time(c.created)

	if v := catalog.GetV(c.OID, ControllerTouched); v != nil {
		if t, ok := v.(time.Time); ok {
			touched = t
		}
	}

	if touched.Before(expired) {
		catalog.PutV(c.OID, ControllerEndpointAddress, nil)
		catalog.PutV(c.OID, ControllerDateTimeCurrent, nil)
		catalog.PutV(c.OID, ControllerCardsCount, nil)
		catalog.PutV(c.OID, ControllerCardsStatus, types.StatusUnknown)
		catalog.PutV(c.OID, ControllerEventsStatus, types.StatusUnknown)
		catalog.PutV(c.OID, ControllerEventsFirst, 0)
		catalog.PutV(c.OID, ControllerEventsLast, 0)
		catalog.PutV(c.OID, ControllerEventsCurrent, 0)

		for _, d := range []uint8{1, 2, 3, 4} {
			if door, ok := c.doors[d]; ok {
				catalog.PutV(door, DoorDelay, nil)
				catalog.PutV(door, DoorControl, nil)
			}
		}

		log.Infof("Controller %v cached values expired", c)

		if c.modified.IsZero() {
			c.deleted = types.TimestampNow()
			catalog.DeleteT(c.CatalogController, c.OID)
			log.Infof("'unconfigured' controller %v removed", c)
		}
	}
}

func (c Controller) serialize() ([]byte, error) {
	if !c.IsValid() || c.IsDeleted() || c.modified.IsZero() {
		return nil, nil
	}

	record := struct {
		OID      schema.OID           `json:"OID,omitempty"`
		Name     string               `json:"name,omitempty"`
		DeviceID uint32               `json:"device-id,omitempty"`
		Address  *core.Address        `json:"address,omitempty"`
		Doors    map[uint8]schema.OID `json:"doors"`
		TimeZone string               `json:"timezone,omitempty"`
		Created  types.Timestamp      `json:"created,omitempty"`
		Modified types.Timestamp      `json:"modified,omitempty"`
	}{
		OID:      c.OID,
		Name:     c.name,
		DeviceID: c.DeviceID,
		Address:  c.IP,
		Doors:    map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""},
		TimeZone: c.timezone,
		Created:  c.created,
		Modified: c.modified,
	}

	for k, v := range c.doors {
		record.Doors[k] = v
	}

	return json.MarshalIndent(record, "", "  ")
}

func (c *Controller) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      schema.OID       `json:"OID"`
		Name     string           `json:"name,omitempty"`
		DeviceID uint32           `json:"device-id,omitempty"`
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone string           `json:"timezone,omitempty"`
		Created  types.Timestamp  `json:"created,omitempty"`
		Modified types.Timestamp  `json:"modified,omitempty"`
	}{
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.name = strings.TrimSpace(record.Name)
	c.DeviceID = record.DeviceID
	c.IP = record.Address
	c.doors = map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""}
	c.timezone = record.TimeZone
	c.created = record.Created
	c.modified = record.Modified

	for k, v := range record.Doors {
		c.doors[k] = schema.OID(v)
	}

	return nil
}

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			CatalogController: catalog.CatalogController{
				OID:      c.OID,
				DeviceID: c.DeviceID,
			},
			name:     c.name,
			IP:       c.IP,
			timezone: c.timezone,
			doors:    map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""},

			created:  c.created,
			modified: c.modified,
			deleted:  c.deleted,
		}

		for k, v := range c.doors {
			replicant.doors[k] = v
		}

		return &replicant
	}

	return nil
}

func (c Controller) updated(uid, field string, before, after interface{}, dbc db.DBC) {
	if dbc != nil {
		description := fmt.Sprintf("Updated %[1]v from %[2]v to %[3]v", field, stringify(before, BLANK), stringify(after, BLANK))

		record := audit.AuditRecord{
			UID:       uid,
			OID:       c.OID,
			Component: "controller",
			Operation: "update",
			Details: audit.Details{
				ID:          stringify(c.DeviceID, ""),
				Name:        stringify(c.name, ""),
				Field:       field,
				Description: description,
				Before:      stringify(before, BLANK),
				After:       stringify(after, BLANK),
			},
		}

		dbc.Write(record)
	}
}

func (c *Controller) log(uid string, operation string, OID schema.OID, field, description, before, after string, dbc db.DBC) {
	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "controller",
		Operation: operation,
		Details: audit.Details{
			ID:          stringify(c.DeviceID, ""),
			Name:        stringify(c.name, ""),
			Field:       field,
			Description: description,
			Before:      before,
			After:       after,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}

func (c icontroller) OID() schema.OID {
	return c.oid
}

func (c icontroller) Name() string {
	return c.name
}

func (c icontroller) ID() uint32 {
	return c.id
}

func (c icontroller) EndPoint() *net.UDPAddr {
	return c.endpoint
}

func (c icontroller) TimeZone() *time.Location {
	return c.timezone
}

func (c icontroller) Door(d uint8) (schema.OID, bool) {
	oid, ok := c.doors[d]

	return oid, ok
}
