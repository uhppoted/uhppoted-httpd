package controllers

import (
	"encoding/json"
	"fmt"
	"math"
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
	OID      catalog.OID           `json:"OID"`
	Name     *types.Name           `json:"name,omitempty"`
	DeviceID *uint32               `json:"device-id,omitempty"`
	IP       *core.Address         `json:"address,omitempty"`
	Doors    map[uint8]catalog.OID `json:"doors"`
	TimeZone *string               `json:"timezone,omitempty"`

	created      time.Time
	deleted      *time.Time
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

const ControllerCreated = catalog.ControllerCreated
const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID
const ControllerAddress = catalog.ControllerAddress
const ControllerAddressConfigured = catalog.ControllerAddressConfigured
const ControllerAddressStatus = catalog.ControllerAddressStatus
const ControllerDateTime = catalog.ControllerDateTime
const ControllerDateTimeSystem = catalog.ControllerDateTimeSystem
const ControllerDateTimeStatus = catalog.ControllerDateTimeStatus
const ControllerCards = catalog.ControllerCards
const ControllerCardsStatus = catalog.ControllerCardsStatus
const ControllerEvents = catalog.ControllerEvents
const ControllerEventsStatus = catalog.ControllerEventsStatus
const ControllerDoor1 = catalog.ControllerDoor1
const ControllerDoor2 = catalog.ControllerDoor2
const ControllerDoor3 = catalog.ControllerDoor3
const ControllerDoor4 = catalog.ControllerDoor4

func (c *Controller) AsObjects() []interface{} {
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
	status := types.StatusUnknown
	name := c.Name
	deviceID := ""
	address := addr{}
	datetime := tinfo{}
	cards := cinfo{}
	events := einfo{}

	doors := map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""}

	if c.DeviceID != nil && *c.DeviceID != 0 {
		deviceID = fmt.Sprintf("%v", *c.DeviceID)
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

	if c.DeviceID != nil && *c.DeviceID != 0 {
		if cached, ok := cache.cache[*c.DeviceID]; ok {
			// ... set status field from cached value
			dt := time.Now().Sub(cached.touched)
			switch {
			case dt < windows.deviceOk:
				status = types.StatusOk

			case dt < windows.deviceUncertain:
				status = types.StatusUncertain
			}

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

	if c.deleted != nil {
		status = types.StatusDeleted
	}

	objects := []interface{}{
		catalog.NewObject(c.OID, status),
		catalog.NewObject2(c.OID, ControllerCreated, created),
		catalog.NewObject2(c.OID, ControllerName, name),
		catalog.NewObject2(c.OID, ControllerDeviceID, deviceID),
		catalog.NewObject2(c.OID, ControllerAddress, address.address),
		catalog.NewObject2(c.OID, ControllerAddressConfigured, address.configured),
		catalog.NewObject2(c.OID, ControllerAddressStatus, address.status),
		catalog.NewObject2(c.OID, ControllerDateTime, datetime.datetime),
		catalog.NewObject2(c.OID, ControllerDateTimeSystem, datetime.system),
		catalog.NewObject2(c.OID, ControllerDateTimeStatus, datetime.status),
		catalog.NewObject2(c.OID, ControllerCards, cards.cards),
		catalog.NewObject2(c.OID, ControllerCardsStatus, cards.status),
		catalog.NewObject2(c.OID, ControllerEvents, events.events),
		catalog.NewObject2(c.OID, ControllerEventsStatus, events.status),
		catalog.NewObject2(c.OID, ControllerDoor1, doors[1]),
		catalog.NewObject2(c.OID, ControllerDoor2, doors[2]),
		catalog.NewObject2(c.OID, ControllerDoor3, doors[3]),
		catalog.NewObject2(c.OID, ControllerDoor4, doors[4]),
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

		if c.DeviceID != nil {
			deviceID = *c.DeviceID
		}

		return &entity{
			Name:     fmt.Sprintf("%v", c.Name),
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
			return c.OID

		case "created":
			return c.created

		case "name":
			return c.Name

		case "id":
			return c.DeviceID
		}
	}

	return nil
}

func (c *Controller) String() string {
	if c == nil {
		return ""
	}

	return fmt.Sprintf("%v (%v)", c.Name, c.DeviceID)
}

func (c *Controller) IsValid() bool {
	if c != nil && (c.Name != nil && *c.Name != "") || (c.DeviceID != nil && *c.DeviceID != 0) {
		return true
	}

	return false
}

func (c *Controller) IsSaveable() bool {
	if c == nil || c.deleted != nil || c.unconfigured {
		return false
	}

	if (c.Name == nil || *c.Name != "") && (c.DeviceID == nil || *c.DeviceID == 0) {
		return false
	}

	return true
}

func (c *Controller) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
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
		case c.OID.Append(ControllerName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					c.OID,
					"name",
					fmt.Sprintf("Updated name from %v to %v", stringify(c.Name, "<blank>"), stringify(value, "<blank>")),
					stringify(c.Name, ""),
					stringify(value, ""),
					dbc)

				name := types.Name(value)
				c.Name = &name
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ControllerName, c.Name))
			}

		case c.OID.Append(ControllerDeviceID):
			if err := f("deviceID", value); err != nil {
				return nil, err
			} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if id, err := strconv.ParseUint(value, 10, 32); err == nil {
					c.log(auth,
						"update",
						c.OID,
						"device-id",
						fmt.Sprintf("Updated device ID from %v to %v", stringify(c.DeviceID, "<blank>"), stringify(value, "<blank>")),
						stringify(c.DeviceID, ""),
						stringify(value, ""),
						dbc)

					cid := uint32(id)
					c.DeviceID = &cid
					c.unconfigured = false
					objects = append(objects, catalog.NewObject2(c.OID, ".2", cid))
				}
			} else if value == "" {
				if p := stringify(c.DeviceID, ""); p != "" {
					c.log(auth,
						"update",
						c.OID,
						"device-id",
						fmt.Sprintf("Cleared device ID %v", p),
						p,
						"",
						dbc)
				} else if p = stringify(c.Name, ""); p != "" {
					c.log(auth,
						"update",
						c.OID,
						"device-id",
						fmt.Sprintf("Cleared device ID for %v", p),
						"",
						"",
						dbc)
				} else {
					c.log(auth,
						"update",
						c.OID,
						"device-id",
						fmt.Sprintf("Cleared device ID"),
						"",
						"",
						dbc)
				}

				c.DeviceID = nil
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".2", ""))
			}

		case c.OID.Append(ControllerAddress):
			if addr, err := core.ResolveAddr(value); err != nil {
				return nil, err
			} else if err := f("address", addr); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					c.OID,
					"address",
					fmt.Sprintf("Updated endpoint from %v to %v", stringify(c.IP, "<blank>"), stringify(value, "<blank>")),
					stringify(c.IP, ""),
					stringify(value, ""),
					dbc)
				c.IP = addr
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".3", c.IP))
			}

		case c.OID.Append(ControllerDateTime):
			if tz, err := types.Timezone(value); err != nil {
				return nil, err
			} else if err := f("timezone", tz); err != nil {
				return nil, err
			} else {
				c.log(auth,
					"update",
					c.OID,
					"timezone",
					fmt.Sprintf("Updated timezone from %v to %v", stringify(c.TimeZone, "<blank>"), stringify(tz.String(), "<blank>")),
					stringify(c.TimeZone, ""),
					stringify(tz.String(), ""),
					dbc)
				tzs := tz.String()
				c.TimeZone = &tzs
				c.unconfigured = false

				if c.DeviceID != nil {
					if cached, ok := cache.cache[*c.DeviceID]; ok {
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
							objects = append(objects, catalog.NewObject2(c.OID, ".4", dt.Format("2006-01-02 15:04 MST")))
						}
					}
				}
			}

		case c.OID.Append(ControllerDoor1):
			if err := f("door[1]", value); err != nil {
				return nil, err
			} else {
				p, _ := catalog.GetV(catalog.OID(c.Doors[1]).Append(catalog.DoorName))
				q, _ := catalog.GetV(catalog.OID(value).Append(catalog.DoorName))
				c.log(auth,
					"update",
					c.OID,
					"door:1",
					fmt.Sprintf("Updated door:1 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[1] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".7", c.Doors[1]))
			}

		case c.OID.Append(ControllerDoor2):
			if err := f("door[2]", value); err != nil {
				return nil, err
			} else {
				p, _ := catalog.GetV(catalog.OID(c.Doors[2]).Append(catalog.DoorName))
				q, _ := catalog.GetV(catalog.OID(value).Append(catalog.DoorName))
				c.log(auth,
					"update",
					c.OID,
					"door:2",
					fmt.Sprintf("Updated door:2 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[2] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".8", c.Doors[2]))
			}

		case c.OID.Append(ControllerDoor3):
			if err := f("door[3]", value); err != nil {
				return nil, err
			} else {
				p, _ := catalog.GetV(catalog.OID(c.Doors[3]).Append(catalog.DoorName))
				q, _ := catalog.GetV(catalog.OID(value).Append(catalog.DoorName))
				c.log(auth,
					"update",
					c.OID,
					"door:3",
					fmt.Sprintf("Updated door:3 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[3] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".9", c.Doors[3]))
			}

		case c.OID.Append(ControllerDoor4):
			if err := f("door[4]", value); err != nil {
				return nil, err
			} else {
				p, _ := catalog.GetV(catalog.OID(c.Doors[4]).Append(catalog.DoorName))
				q, _ := catalog.GetV(catalog.OID(value).Append(catalog.DoorName))
				c.log(auth,
					"update",
					c.OID,
					"door:4",
					fmt.Sprintf("Updated door:4 from %v to %v", stringify(p, "<none>"), stringify(q, "<none>")),
					stringify(p, ""),
					stringify(q, ""),
					dbc)

				c.Doors[4] = catalog.OID(value)
				c.unconfigured = false
				objects = append(objects, catalog.NewObject2(c.OID, ".10", c.Doors[4]))
			}
		}

		if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
			if auth != nil {
				if err := auth.CanDeleteController(c); err != nil {
					return nil, err
				}
			}

			if p := stringify(clone.Name, ""); p != "" {
				clone.log(auth,
					"delete",
					c.OID,
					"device-id",
					fmt.Sprintf("Deleted controller %v", p),
					"",
					"",
					dbc)
			} else if p = stringify(clone.DeviceID, ""); p != "" {
				clone.log(auth,
					"delete",
					c.OID,
					"device-id",
					fmt.Sprintf("Deleted controller %v", p),
					"",
					"",
					dbc)
			} else {
				clone.log(auth,
					"delete",
					c.OID,
					"device-id",
					fmt.Sprintf("Deleted controller"),
					"",
					"",
					dbc)
			}

			now := time.Now()
			c.deleted = &now
			objects = append(objects, catalog.NewObject(c.OID, "deleted"))

			catalog.Delete(c.OID)
		}
	}

	return objects, nil
}

func (c *Controller) deserialize(bytes []byte) error {
	record := struct {
		OID      catalog.OID      `json:"OID"`
		Name     *types.Name      `json:"name,omitempty"`
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
		Created  time.Time        `json:"created"`
	}{}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.Name = record.Name
	c.DeviceID = record.DeviceID
	c.IP = record.Address
	c.Doors = map[uint8]catalog.OID{1: "", 2: "", 3: "", 4: ""}
	c.TimeZone = record.TimeZone
	c.created = record.Created

	for k, v := range record.Doors {
		c.Doors[k] = catalog.OID(v)
	}

	return nil
}

func (c *Controller) serialize() ([]byte, error) {
	if c == nil || c.deleted != nil || c.unconfigured {
		return nil, nil
	}

	if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
		return nil, nil
	}

	record := struct {
		OID      catalog.OID           `json:"OID"`
		Name     *types.Name           `json:"name,omitempty"`
		DeviceID *uint32               `json:"device-id,omitempty"`
		Address  *core.Address         `json:"address,omitempty"`
		Doors    map[uint8]catalog.OID `json:"doors"`
		TimeZone *string               `json:"timezone,omitempty"`
		Created  time.Time             `json:"created"`
	}{
		OID:      c.OID,
		Name:     c.Name,
		DeviceID: c.DeviceID,
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

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			OID:      c.OID,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
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

func (c Controller) stash() {
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
			ID:          stringify(c.DeviceID, ""),
			Name:        stringify(c.Name, ""),
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
