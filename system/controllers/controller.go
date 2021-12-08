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
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	oid      catalog.OID
	name     string
	deviceID *uint32
	IP       *core.Address
	Doors    map[uint8]catalog.OID
	TimeZone *string

	created      types.DateTime
	deleted      *types.DateTime
	unconfigured bool
}

type controller struct {
	OID      catalog.OID
	Name     string
	DeviceID string
	IP       ip
	Doors    map[uint8]string

	Status     types.Status
	SystemTime datetime
	Cards      cards
	Events     *records
	Deleted    bool

	created time.Time
}

type cached struct {
	touched  time.Time
	address  *core.Address
	datetime *types.DateTime
	cards    *uint32
	events   *uint32
	acl      types.Status
}

var created = types.DateTimeNow()

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
	if c != nil && c.deviceID != nil {
		return uint32(*c.deviceID)
	}

	return 0
}

func (c *Controller) EndPoint() *net.UDPAddr {
	if c != nil && c.IP != nil {
		return (*net.UDPAddr)(c.IP)
	}

	return nil
}

func (c *Controller) Door(d uint8) (catalog.OID, bool) {
	if c != nil {
		if v, ok := c.Doors[d]; ok {
			return v, true
		}
	}

	return "", false
}

func (c *Controller) AsObjects() []interface{} {
	OID := c.OID()

	if c.deleted != nil {
		return []interface{}{
			catalog.NewObject2(OID, ControllerDeleted, c.deleted),
		}
	}

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
		events string
		status types.Status
	}

	created := c.created.Format("2006-01-02 15:04:05")
	name := c.name
	deviceID := ""
	address := addr{}
	datetime := tinfo{}
	cards := cinfo{}
	events := einfo{}

	doors := map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""}

	if c.deviceID != nil && *c.deviceID != 0 {
		deviceID = fmt.Sprintf("%v", *c.deviceID)
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

	if c.deviceID != nil && *c.deviceID != 0 {
		//if cached, ok := cache.cache[*c.deviceID]; ok {
		if cached := c.get(); cached != nil {
			// ... set IP address field from cached value
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

			// ... set system date/time field from cached value
			if cached.datetime != nil {
				tz := time.Local
				if c.TimeZone != nil {
					if l, err := timezone(*c.TimeZone); err != nil {
						warn(err)
					} else {
						tz = l
					}
				}

				now := time.Now().In(tz)
				t := time.Time(*cached.datetime)
				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)

				datetime.datetime = T.Format("2006-01-02 15:04:05 MST")
				datetime.system = now.Format("2006-01-02 15:04:05 MST")

				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())
				if delta <= math.Abs(windows.systime.Seconds()) {
					datetime.status = types.StatusOk
				} else {
					datetime.status = types.StatusError
				}
			}

			// ... set ACL field from cached value
			if cached.cards != nil {
				cards.cards = fmt.Sprintf("%d", *cached.cards)
				if cached.acl == types.StatusUnknown {
					cards.status = types.StatusUncertain
				} else {
					cards.status = cached.acl
				}
			}

			// ... set events field from cached value
			events.events = fmt.Sprintf("%v", (*records)(cached.events))
			events.status = types.StatusOk
		}
	}

	objects := []interface{}{
		catalog.NewObject(OID, ""),
		catalog.NewObject2(OID, ControllerStatus, c.status()),
		catalog.NewObject2(OID, ControllerCreated, created),
		catalog.NewObject2(OID, ControllerDeleted, c.deleted),

		catalog.NewObject2(OID, ControllerName, name),
		catalog.NewObject2(OID, ControllerDeviceID, deviceID),
		catalog.NewObject2(OID, ControllerEndpointStatus, address.status),
		catalog.NewObject2(OID, ControllerEndpointAddress, address.address),
		catalog.NewObject2(OID, ControllerEndpointConfigured, address.configured),
		catalog.NewObject2(OID, ControllerDateTimeStatus, datetime.status),
		catalog.NewObject2(OID, ControllerDateTimeCurrent, datetime.datetime),
		catalog.NewObject2(OID, ControllerDateTimeSystem, datetime.system),
		catalog.NewObject2(OID, ControllerCardsStatus, cards.status),
		catalog.NewObject2(OID, ControllerCardsCount, cards.cards),
		catalog.NewObject2(OID, ControllerEventsStatus, events.status),
		catalog.NewObject2(OID, ControllerEventsCount, events.events),
		catalog.NewObject2(OID, ControllerDoor1, doors[1]),
		catalog.NewObject2(OID, ControllerDoor2, doors[2]),
		catalog.NewObject2(OID, ControllerDoor3, doors[3]),
		catalog.NewObject2(OID, ControllerDoor4, doors[4]),
	}

	return objects
}

func (c *Controller) AsRuleEntity() interface{} {
	type entity struct {
		Name     string
		DeviceID uint32
	}

	if c != nil {
		deviceID := uint32(0)

		if c.deviceID != nil {
			deviceID = *c.deviceID
		}

		return &entity{
			Name:     fmt.Sprintf("%v", c.name),
			DeviceID: deviceID,
		}
	}

	return &entity{}
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

	return fmt.Sprintf("%v (%v)", c.name, c.deviceID)
}

func (c *Controller) IsValid() bool {
	if c != nil && (c.name != "" || (c.deviceID != nil && *c.deviceID != 0)) {
		return true
	}

	return false
}

func (c *Controller) IsSaveable() bool {
	if c == nil || c.deleted != nil || c.unconfigured {
		return false
	}

	if c.name != "" && (c.deviceID == nil || *c.deviceID == 0) {
		return false
	}

	return true
}

func (c *Controller) status() types.Status {
	if c.deleted != nil {
		return types.StatusDeleted
	}

	if c.deviceID != nil && *c.deviceID != 0 {
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
			e.datetime = &datetime
		}
	}

	if v := catalog.GetV(c.oid, ControllerCardsCount); v != nil {
		if cards, ok := v.(uint32); ok {
			e.cards = &cards
		}
	}

	if v := catalog.GetV(c.oid, ControllerEventsCount); v != nil {
		if events, ok := v.(*uint32); ok {
			e.events = events
		}
	}

	if v := catalog.GetV(c.oid, ControllerCardsStatus); v != nil {
		if acl, ok := v.(types.Status); ok {
			e.acl = acl
		}
	}

	return &e
}

func (c *Controller) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	OID := c.OID()
	objects := []catalog.Object{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateController(c, field, value)
	}

	if c != nil {
		clone := c.clone()
		switch oid {
		case OID.Append(ControllerName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					OID,
					"name",
					fmt.Sprintf("Updated name from %v to %v", stringify(c.name, BLANK), stringify(value, BLANK)),
					stringify(c.name, ""),
					stringify(value, ""),
					dbc)

				c.name = strings.TrimSpace(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ControllerName, c.name))
			}

		case OID.Append(ControllerDeviceID):
			if err := f("deviceID", value); err != nil {
				return nil, err
			} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if id, err := strconv.ParseUint(value, 10, 32); err == nil {
					c.log(auth,
						"update",
						OID,
						"device-id",
						fmt.Sprintf("Updated device ID from %v to %v", stringify(c.deviceID, BLANK), stringify(value, BLANK)),
						stringify(c.deviceID, ""),
						stringify(value, ""),
						dbc)

					cid := uint32(id)
					c.deviceID = &cid
					c.unconfigured = false
					objects = append(objects, catalog.NewObject2(OID, ".2", cid))
				}
			} else if value == "" {
				if p := stringify(c.deviceID, ""); p != "" {
					c.log(auth,
						"update",
						OID,
						"device-id",
						fmt.Sprintf("Cleared device ID %v", p),
						p,
						"",
						dbc)
				} else if p = stringify(c.name, ""); p != "" {
					c.log(auth,
						"update",
						OID,
						"device-id",
						fmt.Sprintf("Cleared device ID for %v", p),
						"",
						"",
						dbc)
				} else {
					c.log(auth,
						"update",
						OID,
						"device-id",
						fmt.Sprintf("Cleared device ID"),
						"",
						"",
						dbc)
				}

				c.deviceID = nil
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ".2", ""))
			}

		case OID.Append(ControllerEndpointAddress):
			if addr, err := core.ResolveAddr(value); err != nil {
				return nil, err
			} else if err := f("address", addr); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					OID,
					"address",
					fmt.Sprintf("Updated endpoint from %v to %v", stringify(c.IP, BLANK), stringify(value, BLANK)),
					stringify(c.IP, ""),
					stringify(value, ""),
					dbc)
				c.IP = addr
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ".3", c.IP))
			}

		case OID.Append(ControllerDateTimeCurrent):
			if tz, err := types.Timezone(value); err != nil {
				return nil, err
			} else if err := f("timezone", tz); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					OID,
					"timezone",
					fmt.Sprintf("Updated timezone from %v to %v", stringify(c.TimeZone, BLANK), stringify(tz.String(), BLANK)),
					stringify(c.TimeZone, ""),
					stringify(tz.String(), ""),
					dbc)
				tzs := tz.String()
				c.TimeZone = &tzs
				c.unconfigured = false

				if c.deviceID != nil {
					if cached := c.get(); cached != nil {
						if cached.datetime != nil {
							tz := time.Local
							if c.TimeZone != nil {
								if l, err := timezone(*c.TimeZone); err != nil {
									warn(err)
								} else {
									tz = l
								}
							}

							t := time.Time(*cached.datetime)
							dt := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
							objects = append(objects, catalog.NewObject2(OID, ".4", dt.Format("2006-01-02 15:04 MST")))
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
				c.log(auth,
					"update",
					OID,
					"door:1",
					fmt.Sprintf("Updated door:1 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[1] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ControllerDoor1, c.Doors[1]))
			}

		case OID.Append(ControllerDoor2):
			if err := f("door[2]", value); err != nil {
				return nil, err
			} else {
				p := catalog.GetV(catalog.OID(c.Doors[2]), catalog.DoorName)
				q := catalog.GetV(catalog.OID(value), catalog.DoorName)
				c.log(auth,
					"update",
					OID,
					"door:2",
					fmt.Sprintf("Updated door:2 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[2] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ControllerDoor2, c.Doors[2]))
			}

		case OID.Append(ControllerDoor3):
			if err := f("door[3]", value); err != nil {
				return nil, err
			} else {
				p := catalog.GetV(catalog.OID(c.Doors[3]), catalog.DoorName)
				q := catalog.GetV(catalog.OID(value), catalog.DoorName)
				c.log(auth,
					"update",
					OID,
					"door:3",
					fmt.Sprintf("Updated door:3 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[3] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ControllerDoor3, c.Doors[3]))
			}

		case OID.Append(ControllerDoor4):
			if err := f("door[4]", value); err != nil {
				return nil, err
			} else {
				p := catalog.GetV(catalog.OID(c.Doors[4]), catalog.DoorName)
				q := catalog.GetV(catalog.OID(value), catalog.DoorName)
				c.log(auth,
					"update",
					OID,
					"door:4",
					fmt.Sprintf("Updated door:4 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[4] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(OID, ControllerDoor4, c.Doors[4]))
			}
		}

		if c.name == "" && (c.deviceID == nil || *c.deviceID == 0) {
			if auth != nil {
				if err := auth.CanDeleteController(c); err != nil {
					return nil, err
				}
			}

			if p := stringify(clone.name, ""); p != "" {
				clone.log(auth,
					"delete",
					OID,
					"device-id",
					fmt.Sprintf("Deleted controller %v", p),
					"",
					"",
					dbc)
			} else if p = stringify(clone.deviceID, ""); p != "" {
				clone.log(auth,
					"delete",
					OID,
					"device-id",
					fmt.Sprintf("Deleted controller %v", p),
					"",
					"",
					dbc)
			} else {
				clone.log(auth,
					"delete",
					OID,
					"device-id",
					fmt.Sprintf("Deleted controller"),
					"",
					"",
					dbc)
			}

			now := types.DateTime(time.Now())
			c.deleted = &now
			objects = append(objects, catalog.NewObject(OID, "deleted"))
			objects = append(objects, catalog.NewObject2(OID, ControllerDeleted, c.deleted))

			catalog.Delete(OID)
		}

		objects = append(objects, catalog.NewObject2(OID, ControllerStatus, c.status()))
		objects = append(objects, catalog.NewObject2(OID, ControllerDeleted, c.deleted))
	}

	return objects, nil
}

func (c *Controller) serialize() ([]byte, error) {
	if c == nil || c.deleted != nil || c.unconfigured {
		return nil, nil
	}

	if c.name == "" && (c.deviceID == nil || *c.deviceID == 0) {
		return nil, nil
	}

	record := struct {
		OID      catalog.OID           `json:"OID,omitempty"`
		Name     string                `json:"name,omitempty"`
		DeviceID *uint32               `json:"device-id,omitempty"`
		Address  *core.Address         `json:"address,omitempty"`
		Doors    map[uint8]catalog.OID `json:"doors"`
		TimeZone *string               `json:"timezone,omitempty"`
		Created  types.DateTime        `json:"created"`
	}{
		OID:      c.OID(),
		Name:     c.name,
		DeviceID: c.deviceID,
		Address:  c.IP,
		Doors:    map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""},
		TimeZone: c.TimeZone,
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
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
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
	c.TimeZone = record.TimeZone
	c.created = record.Created

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
			TimeZone: c.TimeZone,
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
