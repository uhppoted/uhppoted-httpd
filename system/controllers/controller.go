package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	lib "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	catalog.CatalogController
	name         string
	IP           lib.ControllerAddr
	doors        map[uint8]schema.OID
	interlock    lib.Interlock
	antipassback lib.AntiPassback
	timezone     string
	protocol     string

	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp
}

type icontroller struct {
	oid          schema.OID
	name         string
	id           uint32
	endpoint     lib.ControllerAddr
	timezone     *time.Location
	protocol     string
	doors        map[uint8]schema.OID
	interlock    lib.Interlock
	antipassback lib.AntiPassback
	synchronized types.Status
}

type kv = struct {
	field schema.Suffix
	value any
}

type cached struct {
	touched      time.Time
	address      lib.ControllerAddr
	protocol     string
	antipassback lib.AntiPassback
	datetime     struct {
		datetime lib.DateTime
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
		return fmt.Errorf("at least one of controller name and device ID must be valid")
	}

	return nil
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
			datetime   string
			configured string
			status     types.Status
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
		interlock := ""
		antipassback := ""

		if c.DeviceID != 0 {
			deviceID = fmt.Sprintf("%v", c.DeviceID)
		}

		if c.IP.IsValid() {
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
				if cached.address.IsValid() {
					address.address = fmt.Sprintf("%v", cached.address)
					switch {
					case !c.IP.IsValid() || (c.IP.IsValid() && cached.address.Equal(c.IP)):
						address.status = types.StatusOk

					case c.IP.IsValid() && !cached.address.Equal(c.IP):
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

					if tz.String() == time.Local.String() {
						datetime.datetime = T.Format("2006-01-02 15:04:05")
						datetime.configured = now.Format("2006-01-02 15:04:05")
					} else {
						datetime.datetime = T.Format("2006-01-02 15:04:05 MST")
						datetime.configured = now.Format("2006-01-02 15:04:05 MST")
					}

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

				// ... get anti-passback field from cached value
				antipassback = fmt.Sprintf("%v", uint8(cached.antipassback))
			}
		}

		if c.DeviceID != 0 {
			interlock = fmt.Sprintf("%v", uint8(c.interlock))
		}

		list = append(list, kv{ControllerStatus, c.Status()})
		list = append(list, kv{ControllerCreated, c.created})
		list = append(list, kv{ControllerDeleted, c.deleted})
		list = append(list, kv{ControllerName, name})
		list = append(list, kv{ControllerDeviceID, deviceID})
		list = append(list, kv{ControllerEndpointStatus, address.status})
		list = append(list, kv{ControllerEndpointAddress, address.address})
		list = append(list, kv{ControllerEndpointProtocol, c.protocol})
		list = append(list, kv{ControllerEndpointConfigured, address.configured})
		list = append(list, kv{ControllerDateTimeStatus, datetime.status})
		list = append(list, kv{ControllerDateTimeCurrent, datetime.datetime})
		list = append(list, kv{ControllerDateTimeConfigured, datetime.configured})
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
		list = append(list, kv{ControllerInterlock, interlock})
		list = append(list, kv{ControllerAntiPassback, antipassback})
		// FIXME list = append(list, kv{ControllerAntiPassback, antipassback.antipassback})
		// FIXME list = append(list, kv{ControllerAntiPassbackConfigured, antipassback.configured})
	}

	return c.toObjects(list, a)
}

func (c Controller) AsRuleEntity() (string, interface{}) {
	v := struct {
		Name     string
		DeviceID uint32
	}{
		Name:     c.name,
		DeviceID: c.DeviceID,
	}

	return "controller", &v
}

func (c Controller) CacheKey() string {
	return ""
}

func (c *Controller) AsIController() types.IController {
	var endpoint lib.ControllerAddr
	var location *time.Location = time.Local
	var doors = map[uint8]schema.OID{}

	if c.IP.IsValid() {
		endpoint = c.IP
	}

	if tz, err := types.Timezone(c.timezone); err == nil && tz != nil {
		location = tz
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.doors[d]; ok {
			doors[d] = oid
		}
	}

	// ... get system date/time synchronization
	synchronized := types.StatusUnknown
	if cached := c.get(); cached != nil && !cached.datetime.datetime.IsZero() && !cached.datetime.modified {
		tz, err := types.Timezone(c.timezone)
		if err != nil {
			tz = time.Local
		}

		t := time.Time(cached.datetime.datetime)
		T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
		delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

		if delta <= math.Abs(windows.systime.Seconds()) {
			synchronized = types.StatusOk
		} else {
			synchronized = types.StatusError
		}
	}

	return &icontroller{
		oid:          c.OID,
		name:         c.name,
		id:           c.DeviceID,
		endpoint:     endpoint,
		timezone:     location,
		protocol:     c.protocol,
		doors:        doors,
		interlock:    c.interlock,
		antipassback: c.antipassback,
		synchronized: synchronized,
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
			dt := time.Since(cached.touched)
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
		if address, ok := v.(lib.ControllerAddr); ok {
			e.address = address
		}
	}

	if v := catalog.GetV(c.OID, ControllerEndpointProtocol); v != nil {
		if protocol, ok := v.(string); ok {
			e.protocol = protocol
		}
	}

	if v := catalog.GetV(c.OID, ControllerDateTimeCurrent); v != nil {
		if datetime, ok := v.(lib.DateTime); ok {
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

	if v := catalog.GetV(c.OID, ControllerAntiPassback); v != nil {
		if antipassback, ok := v.(lib.AntiPassback); ok {
			e.antipassback = antipassback
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

	uid := auth.UID(a)
	clone := c.clone()
	list := []kv{}

	switch oid {
	case c.OID.Append(ControllerName):
		if err := CanUpdate(a, c, "name", value); err != nil {
			return nil, err
		} else {
			c.name = strings.TrimSpace(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerName, c.name})

			c.updated(dbc, uid, "name", clone.name, c.name)
		}

	case c.OID.Append(ControllerDeviceID):
		if err := CanUpdate(a, c, "deviceID", value); err != nil {
			return nil, err
		} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
			if id, err := strconv.ParseUint(value, 10, 32); err == nil {
				c.DeviceID = uint32(id)
				c.modified = types.TimestampNow()

				list = append(list, kv{ControllerDeviceID, c.DeviceID})
				c.updated(dbc, uid, "device-id", clone.DeviceID, c.DeviceID)
			}
		} else if value == "" {
			if p := fmt.Sprintf("%v", types.Uint32(c.DeviceID)); p != "" {
				c.log(dbc, uid, "update", "device-id", p, "", "Cleared device ID %v", p)
			} else if c.name != "" {
				c.log(dbc, uid, "update", "device-id", "", "", "Cleared device ID for %v", c.name)
			} else {
				c.log(dbc, uid, "update", "device-id", "", "", "Cleared device ID")
			}

			c.DeviceID = 0
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDeviceID, ""})
		}

	case c.OID.Append(ControllerEndpointAddress):
		if addr, err := lib.ParseControllerAddr(value); err != nil {
			return nil, err
		} else if err := CanUpdate(a, c, "address", addr); err != nil {
			return nil, err
		} else {
			c.IP = addr
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerEndpointAddress, addr})
			list = append(list, kv{ControllerEndpointConfigured, addr})
			list = append(list, kv{ControllerEndpointStatus, types.StatusUncertain})

			c.updated(dbc, uid, "address", clone.IP, c.IP)
		}

	case c.OID.Append(ControllerEndpointProtocol):
		if err := CanUpdate(a, c, "protocol", value); err != nil {
			return nil, err
		} else {
			c.protocol = value
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerEndpointProtocol, value})

			c.updated(dbc, uid, "protocol", clone.protocol, c.protocol)
		}

	case c.OID.Append(ControllerDateTimeCurrent):
		if tz, err := types.Timezone(value); err != nil {
			return nil, err
		} else if err := CanUpdate(a, c, "timezone", tz); err != nil {
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
						list = append(list, kv{ControllerDateTimeConfigured, dt.Format("2006-01-02 15:04 MST")})
						list = append(list, kv{ControllerDateTimeModified, true})

						if tz.String() == time.Local.String() {
							list = append(list, kv{ControllerDateTimeConfigured, dt.Format("2006-01-02 15:04")})
						} else {
							list = append(list, kv{ControllerDateTimeConfigured, dt.Format("2006-01-02 15:04 MST")})
						}

						dbc.Updated(c.OID, ControllerDateTime, dt)
					}
				}
			}

			c.updated(dbc, uid, "timezone", clone.timezone, c.timezone)
		}

	case c.OID.Append(ControllerDoor1):
		if err := CanUpdate(a, c, "door[1]", value); err != nil {
			return nil, err
		} else {
			c.doors[1] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor1, c.doors[1]})

			c.updated(dbc, uid, "door:1", clone.doors[1], c.doors[1])
		}

	case c.OID.Append(ControllerDoor2):
		if err := CanUpdate(a, c, "door[2]", value); err != nil {
			return nil, err
		} else {
			c.doors[2] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor2, c.doors[2]})

			c.updated(dbc, uid, "door:2", clone.doors[2], c.doors[2])
		}

	case c.OID.Append(ControllerDoor3):
		if err := CanUpdate(a, c, "door[3]", value); err != nil {
			return nil, err
		} else {
			c.doors[3] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor3, c.doors[3]})

			c.updated(dbc, uid, "door:3", clone.doors[3], c.doors[3])
		}

	case c.OID.Append(ControllerDoor4):
		if err := CanUpdate(a, c, "door[4]", value); err != nil {
			return nil, err
		} else {
			c.doors[4] = schema.OID(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerDoor4, c.doors[4]})

			c.updated(dbc, uid, "door:4", clone.doors[4], c.doors[4])
		}

	case c.OID.Append(ControllerInterlock):
		if err := CanUpdate(a, c, "interlock", value); err != nil {
			return nil, err
		} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
			return nil, err
		} else if v != 0 && v != 1 && v != 2 && v != 3 && v != 4 && v != 8 {
			return nil, fmt.Errorf("invalid interlock mode (%v)", v)
		} else {
			c.interlock = lib.Interlock(v)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerInterlock, uint8(c.interlock)})

			dbc.Updated(c.OID, ControllerInterlock, c.interlock)
			c.updated(dbc, uid, "interlock", clone.interlock, c.interlock)
		}

	case c.OID.Append(ControllerAntiPassback):
		if err := CanUpdate(a, c, "antipassback", value); err != nil {
			return nil, err
		} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
			return nil, err
		} else if v != 0 && v != 1 && v != 2 && v != 3 && v != 4 {
			return nil, fmt.Errorf("invalid antipassback mode (%v)", v)
		} else {
			c.antipassback = lib.AntiPassback(v)
			c.modified = types.TimestampNow()

			list = append(list, kv{ControllerAntiPassback, uint8(c.antipassback)})

			dbc.Updated(c.OID, ControllerAntiPassback, c.antipassback)
			c.updated(dbc, uid, "antipassback", clone.antipassback, c.antipassback)
		}
	}

	list = append(list, kv{ControllerStatus, c.Status()})

	return c.toObjects(list, a), nil
}

func (c *Controller) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if c != nil {
		uid := auth.UID(a)

		if err := CanDelete(a, c); err != nil {
			return nil, err
		}

		if c.name != "" {
			c.log(dbc, uid, "delete", "device-id", "", "", "Deleted controller %v", c.name)
		} else if s := fmt.Sprintf("%v", types.Uint32(c.DeviceID)); s != "" {
			c.log(dbc, uid, "delete", "device-id", "", "", "Deleted controller %v", s)
		} else {
			c.log(dbc, uid, "delete", "device-id", "", "", "Deleted controller")
		}

		c.deleted = types.TimestampNow()
		c.modified = types.TimestampNow()

		list = append(list, kv{ControllerDeleted, c.deleted})
		list = append(list, kv{ControllerStatus, c.Status()})

		catalog.DeleteT(c.CatalogController, c.OID)
	}

	return c.toObjects(list, a), nil
}

func (c Controller) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	OID := c.OID
	objects := []schema.Object{}

	if err := CanView(a, c, "OID", OID); err == nil && !c.IsDeleted() {
		catalog.Join(&objects, catalog.NewObject(OID, ""))
	}

	for _, v := range list {
		field := lookup[v.field]
		if err := CanView(a, c, field, v.value); err == nil {
			catalog.Join(&objects, catalog.NewObject2(OID, v.field, v.value))
		}
	}

	return objects
}

func (c Controller) serialize() ([]byte, error) {
	if !c.IsValid() || c.IsDeleted() || c.modified.IsZero() {
		return nil, nil
	}

	record := struct {
		OID          schema.OID           `json:"OID,omitempty"`
		Name         string               `json:"name,omitempty"`
		DeviceID     uint32               `json:"device-id,omitempty"`
		Address      lib.ControllerAddr   `json:"address,omitempty"`
		Doors        map[uint8]schema.OID `json:"doors"`
		Interlock    lib.Interlock        `json:"interlock"`
		AntiPassback lib.AntiPassback     `json:"antipassback"`
		TimeZone     string               `json:"timezone,omitempty"`
		Protocol     string               `json:"protocol,omitempty"`
		Created      types.Timestamp      `json:"created,omitempty"`
		Modified     types.Timestamp      `json:"modified,omitempty"`
	}{
		OID:          c.OID,
		Name:         c.name,
		DeviceID:     c.DeviceID,
		Address:      c.IP,
		Doors:        map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""},
		Interlock:    c.interlock,
		AntiPassback: c.antipassback,
		TimeZone:     c.timezone,
		Protocol:     c.protocol,
		Created:      c.created.UTC(),
		Modified:     c.modified.UTC(),
	}

	for k, v := range c.doors {
		record.Doors[k] = v
	}

	return json.MarshalIndent(record, "", "  ")
}

func (c *Controller) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID          schema.OID         `json:"OID"`
		Name         string             `json:"name,omitempty"`
		DeviceID     uint32             `json:"device-id,omitempty"`
		Address      lib.ControllerAddr `json:"address,omitempty"`
		Doors        map[uint8]string   `json:"doors"`
		Interlock    lib.Interlock      `json:"interlock"`
		AntiPassback lib.AntiPassback   `json:"antipassback"`
		TimeZone     string             `json:"timezone,omitempty"`
		Protocol     string             `json:"protocol,omitempty"`
		Created      types.Timestamp    `json:"created,omitempty"`
		Modified     types.Timestamp    `json:"modified,omitempty"`
	}{
		Created:  created,
		Protocol: "udp",
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.name = strings.TrimSpace(record.Name)
	c.DeviceID = record.DeviceID
	c.IP = record.Address
	c.doors = map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""}
	c.interlock = record.Interlock
	c.antipassback = record.AntiPassback
	c.timezone = record.TimeZone
	c.protocol = record.Protocol
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
			name:         c.name,
			IP:           c.IP,
			timezone:     c.timezone,
			protocol:     c.protocol,
			doors:        map[uint8]schema.OID{1: "", 2: "", 3: "", 4: ""},
			interlock:    c.interlock,
			antipassback: c.antipassback,

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

func (c Controller) updated(dbc db.DBC, uid, field string, before, after interface{}) {
	c.log(dbc, uid, "update", field, before, after, "Updated %[1]v from '%[2]v' to '%[3]v'", field, before, after)
}

func (c *Controller) log(dbc db.DBC, uid, op string, field string, before, after any, format string, fields ...any) {
	dbc.Log(uid, op, c.OID, "controller", c.DeviceID, c.name, field, before, after, format, fields...)
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

func (c icontroller) EndPoint() lib.ControllerAddr {
	return c.endpoint
}

func (c icontroller) TimeZone() *time.Location {
	return c.timezone
}

func (c icontroller) Protocol() string {
	if c.protocol == "tcp" {
		return "tcp"
	}

	return "udp"
}

func (c icontroller) Door(d uint8) (schema.OID, bool) {
	oid, ok := c.doors[d]

	return oid, ok
}

func (c icontroller) DateTimeOk() bool {
	return c.synchronized != types.StatusError
}
