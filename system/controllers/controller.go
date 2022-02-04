package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
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

type Controller struct {
	oid      catalog.OID
	name     string
	deviceID uint32
	IP       *core.Address
	Doors    map[uint8]catalog.OID
	timezone string

	created      types.DateTime
	deleted      types.DateTime
	unconfigured bool
}

type kv = struct {
	field catalog.Suffix
	value interface{}
}

type cached struct {
	touched  time.Time
	address  *core.Address
	datetime struct {
		datetime *types.DateTime
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

var created = types.DateTimeNow()

func (c Controller) IsDeleted() bool {
	return !c.deleted.IsZero()
}

func (c *Controller) OID() catalog.OID {
	if c != nil {
		return c.oid
	}

	return ""
}

func (c *Controller) Name() string {
	if c != nil {
		return c.name
	}

	return ""
}

func (c *Controller) DeviceID() uint32 {
	if c != nil {
		return c.deviceID
	}

	return 0
}

func (c *Controller) EndPoint() *net.UDPAddr {
	if c != nil && c.IP != nil {
		return (*net.UDPAddr)(c.IP)
	}

	return nil
}

func (c *Controller) TimeZone() *time.Location {
	location := time.Local
	if tz, err := types.Timezone(c.timezone); err == nil && tz != nil {
		location = tz
	}

	return location
}

func (c *Controller) Door(d uint8) (catalog.OID, bool) {
	if c != nil {
		if v, ok := c.Doors[d]; ok {
			return v, true
		}
	}

	return "", false
}

func (c *Controller) realized() bool {
	if c != nil && c.deviceID != 0 && !c.IsDeleted() {
		return true
	}

	return false
}

func (c *Controller) AsObjects(auth auth.OpAuth) []catalog.Object {
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

		doors := map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""}

		if c.deviceID != 0 {
			deviceID = fmt.Sprintf("%v", c.deviceID)
		}

		if c.IP != nil {
			address.address = fmt.Sprintf("%v", c.IP)
			address.configured = fmt.Sprintf("%v", c.IP)
		}

		for _, i := range []uint8{1, 2, 3, 4} {
			if d, ok := c.Doors[i]; ok {
				doors[i] = d
			}
		}

		if c.deviceID != 0 {
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
				if cached.datetime.datetime != nil {
					tz := timezone(c.timezone)
					now := time.Now().In(tz)
					t := time.Time(*cached.datetime.datetime)
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

		list = append(list, kv{ControllerStatus, c.status()})
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

	return c.toObjects(list, auth)
}

func (c *Controller) AsRuleEntity() (string, interface{}) {
	v := struct {
		Name     string
		DeviceID uint32
	}{}

	if c != nil {
		v.Name = c.name
		v.DeviceID = c.deviceID
	}

	return "controller", &v
}

func (c *Controller) Get(key string) interface{} {
	f := strings.ToLower(key)
	if c != nil {
		switch f {
		case "oid":
			return c.OID()

		case "created":
			return c.created

		case "name":
			return c.name

		case "id":
			return c.deviceID
		}
	}

	return nil
}

func (c *Controller) String() string {
	if c == nil {
		return ""
	}

	if deviceID := c.DeviceID(); deviceID == 0 {
		return fmt.Sprintf("%v", c.Name())
	} else {
		return fmt.Sprintf("%v (%v)", c.Name(), deviceID)
	}
}

func (c *Controller) IsValid() bool {
	if c != nil && (c.name != "" || c.deviceID != 0) {
		return true
	}

	return false
}

func (c *Controller) status() types.Status {
	if c.IsDeleted() {
		return types.StatusDeleted
	}

	if c.deviceID != 0 {
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

	if v := catalog.GetV(c.oid, ControllerTouched); v != nil {
		if touched, ok := v.(time.Time); ok {
			e.touched = touched
		}
	}

	if v := catalog.GetV(c.oid, ControllerEndpointAddress); v != nil {
		if address, ok := v.(core.Address); ok {
			e.address = &address
		}
	}

	if v := catalog.GetV(c.oid, ControllerDateTimeCurrent); v != nil {
		if datetime, ok := v.(types.DateTime); ok {
			e.datetime.datetime = &datetime
		}
	}

	if v := catalog.GetV(c.OID(), ControllerDateTimeModified); v != nil {
		if b, ok := v.(bool); ok {
			e.datetime.modified = b
		}
	}

	if v := catalog.GetV(c.oid, ControllerCardsCount); v != nil {
		if cards, ok := v.(uint32); ok {
			e.cards = &cards
		}
	}

	if v := catalog.GetV(c.oid, ControllerEventsStatus); v != nil {
		if status, ok := v.(types.Status); ok {
			e.events.status = status
		}
	}

	if v := catalog.GetV(c.oid, ControllerEventsFirst); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.first = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.oid, ControllerEventsLast); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.last = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.oid, ControllerEventsCurrent); v != nil {
		if index, ok := v.(uint32); ok {
			e.events.current = types.Uint32(index)
		}
	}

	if v := catalog.GetV(c.oid, ControllerCardsStatus); v != nil {
		if acl, ok := v.(types.Status); ok {
			e.acl = acl
		}
	}

	return &e
}

func (c *Controller) set(a auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if c == nil {
		return []catalog.Object{}, nil
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

	OID := c.OID()
	list := []kv{}
	clone := c.clone()

	switch oid {
	case OID.Append(ControllerName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			c.log(a,
				"update",
				OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(c.name, BLANK), stringify(value, BLANK)),
				stringify(c.name, ""),
				stringify(value, ""),
				dbc)

			c.name = strings.TrimSpace(value)
			c.unconfigured = false
			list = append(list, kv{ControllerName, c.name})
		}

	case OID.Append(ControllerDeviceID):
		if err := f("deviceID", value); err != nil {
			return nil, err
		} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
			if id, err := strconv.ParseUint(value, 10, 32); err == nil {
				c.log(a,
					"update",
					OID,
					"device-id",
					fmt.Sprintf("Updated device ID from %v to %v", stringify(c.deviceID, BLANK), stringify(value, BLANK)),
					stringify(c.deviceID, ""),
					stringify(value, ""),
					dbc)

				c.deviceID = uint32(id)
				c.unconfigured = false
				list = append(list, kv{ControllerDeviceID, c.deviceID})
			}
		} else if value == "" {
			if p := stringify(c.deviceID, ""); p != "" {
				c.log(a, "update", OID, "device-id", fmt.Sprintf("Cleared device ID %v", p), p, "", dbc)
			} else if p = stringify(c.name, ""); p != "" {
				c.log(a, "update", OID, "device-id", fmt.Sprintf("Cleared device ID for %v", p), "", "", dbc)
			} else {
				c.log(a, "update", OID, "device-id", fmt.Sprintf("Cleared device ID"), "", "", dbc)
			}

			c.deviceID = 0
			c.unconfigured = false
			list = append(list, kv{ControllerDeviceID, ""})
		}

	case OID.Append(ControllerEndpointAddress):
		if addr, err := core.ResolveAddr(value); err != nil {
			return nil, err
		} else if err := f("address", addr); err != nil {
			return nil, err
		} else {
			c.log(a,
				"update",
				OID,
				"address",
				fmt.Sprintf("Updated endpoint from %v to %v", stringify(c.IP, BLANK), stringify(value, BLANK)),
				stringify(c.IP, ""),
				stringify(value, ""),
				dbc)
			c.IP = addr
			c.unconfigured = false
			list = append(list, kv{".3", c.IP})
		}

	case OID.Append(ControllerDateTimeCurrent):
		if tz, err := types.Timezone(value); err != nil {
			return nil, err
		} else if err := f("timezone", tz); err != nil {
			return nil, err
		} else {
			c.log(a,
				"update",
				OID,
				"timezone",
				fmt.Sprintf("Updated timezone from %v to %v", stringify(c.timezone, BLANK), stringify(tz.String(), BLANK)),
				stringify(c.timezone, ""),
				stringify(tz.String(), ""),
				dbc)

			c.timezone = tz.String()
			c.unconfigured = false

			if c.deviceID != 0 {
				if cached := c.get(); cached != nil {
					if cached.datetime.datetime != nil {
						tz := timezone(c.timezone)
						t := time.Time(*cached.datetime.datetime)
						dt := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
						list = append(list, kv{ControllerDateTimeStatus, types.StatusUncertain})
						list = append(list, kv{ControllerDateTime, dt.Format("2006-01-02 15:04 MST")})
						list = append(list, kv{ControllerDateTimeModified, true})
					}
				}
			}
		}

	case OID.Append(ControllerDoor1):
		if err := f("door[1]", value); err != nil {
			return nil, err
		} else {
			p := catalog.GetV(catalog.OID(c.Doors[1]), catalog.DoorName)
			q := catalog.GetV(catalog.OID(value), catalog.DoorName)
			c.log(a,
				"update",
				OID,
				"door:1",
				fmt.Sprintf("Updated door:1 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
				stringify(p, ""),
				stringify(q, ""),
				dbc)

			c.Doors[1] = catalog.OID(value)
			c.unconfigured = false
			list = append(list, kv{ControllerDoor1, c.Doors[1]})
		}

	case OID.Append(ControllerDoor2):
		if err := f("door[2]", value); err != nil {
			return nil, err
		} else {
			p := catalog.GetV(catalog.OID(c.Doors[2]), catalog.DoorName)
			q := catalog.GetV(catalog.OID(value), catalog.DoorName)
			c.log(a,
				"update",
				OID,
				"door:2",
				fmt.Sprintf("Updated door:2 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
				stringify(p, ""),
				stringify(q, ""),
				dbc)

			c.Doors[2] = catalog.OID(value)
			c.unconfigured = false
			list = append(list, kv{ControllerDoor2, c.Doors[2]})
		}

	case OID.Append(ControllerDoor3):
		if err := f("door[3]", value); err != nil {
			return nil, err
		} else {
			p := catalog.GetV(catalog.OID(c.Doors[3]), catalog.DoorName)
			q := catalog.GetV(catalog.OID(value), catalog.DoorName)
			c.log(a,
				"update",
				OID,
				"door:3",
				fmt.Sprintf("Updated door:3 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
				stringify(p, ""),
				stringify(q, ""),
				dbc)

			c.Doors[3] = catalog.OID(value)
			c.unconfigured = false
			list = append(list, kv{ControllerDoor3, c.Doors[3]})
		}

	case OID.Append(ControllerDoor4):
		if err := f("door[4]", value); err != nil {
			return nil, err
		} else {
			p := catalog.GetV(catalog.OID(c.Doors[4]), catalog.DoorName)
			q := catalog.GetV(catalog.OID(value), catalog.DoorName)
			c.log(a,
				"update",
				OID,
				"door:4",
				fmt.Sprintf("Updated door:4 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
				stringify(p, ""),
				stringify(q, ""),
				dbc)

			c.Doors[4] = catalog.OID(value)
			c.unconfigured = false
			list = append(list, kv{ControllerDoor4, c.Doors[4]})
		}
	}

	if c.name == "" && c.deviceID == 0 {
		if a != nil {
			if err := a.CanDelete(c, auth.Controllers); err != nil {
				return nil, err
			}
		}

		if p := stringify(clone.name, ""); p != "" {
			clone.log(a,
				"delete",
				OID,
				"device-id",
				fmt.Sprintf("Deleted controller %v", p),
				"",
				"",
				dbc)
		} else if p = stringify(clone.deviceID, ""); p != "" {
			clone.log(a,
				"delete",
				OID,
				"device-id",
				fmt.Sprintf("Deleted controller %v", p),
				"",
				"",
				dbc)
		} else {
			clone.log(a,
				"delete",
				OID,
				"device-id",
				fmt.Sprintf("Deleted controller"),
				"",
				"",
				dbc)
		}

		c.deleted = types.DateTimeNow()
		list = append(list, kv{ControllerDeleted, c.deleted})

		catalog.Delete(OID)
	}

	list = append(list, kv{ControllerStatus, c.status()})

	return c.toObjects(list, a), nil
}

func (c *Controller) toObjects(list []kv, a auth.OpAuth) []catalog.Object {
	f := func(c *Controller, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(c, field, value, auth.Controllers); err != nil {
				return false
			}
		}

		return true
	}

	OID := c.OID()
	objects := []catalog.Object{}

	if !c.IsDeleted() && f(c, "OID", OID) {
		objects = append(objects, catalog.NewObject(OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(c, field, v.value) {
			objects = append(objects, catalog.NewObject2(OID, v.field, v.value))
		}
	}

	return objects
}

func (c *Controller) refreshed() {
	expired := time.Now().Add(-windows.cacheExpiry)

	touched := time.Time(c.created)

	if v := catalog.GetV(c.OID(), ControllerTouched); v != nil {
		if t, ok := v.(time.Time); ok {
			touched = t
		}
	}

	if touched.Before(expired) {
		catalog.PutV(c.OID(), ControllerEndpointAddress, nil)
		catalog.PutV(c.OID(), ControllerDateTimeCurrent, nil)
		catalog.PutV(c.OID(), ControllerCardsCount, nil)
		catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusUnknown)
		catalog.PutV(c.OID(), ControllerEventsStatus, types.StatusUnknown)
		catalog.PutV(c.OID(), ControllerEventsFirst, 0)
		catalog.PutV(c.OID(), ControllerEventsLast, 0)
		catalog.PutV(c.OID(), ControllerEventsCurrent, 0)

		for _, d := range []uint8{1, 2, 3, 4} {
			if door, ok := c.Door(d); ok {
				catalog.PutV(door, DoorDelay, nil)
				catalog.PutV(door, DoorControl, nil)
			}
		}

		log.Printf("Controller %v cached values expired", c)

		if c.unconfigured {
			c.deleted = types.DateTimeNow()
			catalog.Delete(c.OID())
			log.Printf("'unconfigured' controller %v removed", c)
		}
	}
}

func (c Controller) serialize() ([]byte, error) {
	if !c.IsValid() || c.IsDeleted() || c.unconfigured {
		return nil, nil
	}

	record := struct {
		OID      catalog.OID           `json:"OID,omitempty"`
		Name     string                `json:"name,omitempty"`
		DeviceID uint32                `json:"device-id,omitempty"`
		Address  *core.Address         `json:"address,omitempty"`
		Doors    map[uint8]catalog.OID `json:"doors"`
		TimeZone string                `json:"timezone,omitempty"`
		Created  types.DateTime        `json:"created"`
	}{
		OID:      c.OID(),
		Name:     c.name,
		DeviceID: c.deviceID,
		Address:  c.IP,
		Doors:    map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""},
		TimeZone: c.timezone,
		Created:  c.created,
	}

	for k, v := range c.Doors {
		record.Doors[k] = v
	}

	return json.MarshalIndent(record, "", "  ")
}

func (c *Controller) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      catalog.OID      `json:"OID"`
		Name     string           `json:"name,omitempty"`
		DeviceID uint32           `json:"device-id,omitempty"`
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone string           `json:"timezone,omitempty"`
		Created  types.DateTime   `json:"created,omitempty"`
	}{
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.oid = record.OID
	c.name = strings.TrimSpace(record.Name)
	c.deviceID = record.DeviceID
	c.IP = record.Address
	c.Doors = map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""}
	c.timezone = record.TimeZone
	c.created = record.Created
	c.unconfigured = false

	for k, v := range record.Doors {
		c.Doors[k] = catalog.OID(v)
	}

	return nil
}

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			oid:      c.oid,
			name:     c.name,
			deviceID: c.deviceID,
			IP:       c.IP,
			timezone: c.timezone,
			Doors:    map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""},

			created:      c.created,
			deleted:      c.deleted,
			unconfigured: c.unconfigured,
		}

		for k, v := range c.Doors {
			replicant.Doors[k] = v
		}

		return &replicant
	}

	return nil
}

func (c *Controller) log(auth auth.OpAuth, operation string, OID catalog.OID, field, description, before, after string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "controller",
		Operation: operation,
		Details: audit.Details{
			ID:          stringify(c.deviceID, ""),
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
