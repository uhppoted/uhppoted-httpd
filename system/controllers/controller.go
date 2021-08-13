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
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	OID      string           `json:"OID"`
	Name     *types.Name      `json:"name,omitempty"`
	DeviceID *uint32          `json:"device-id,omitempty"`
	IP       *core.Address    `json:"address,omitempty"`
	Doors    map[uint8]string `json:"doors"`
	TimeZone *string          `json:"timezone,omitempty"`

	created      time.Time
	deleted      *time.Time
	unconfigured bool
}

type controller struct {
	OID      string
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

func (c *Controller) deserialize(bytes []byte) error {
	record := struct {
		OID      string           `json:"OID"`
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
	c.Doors = map[uint8]string{1: "", 2: "", 3: "", 4: ""}
	c.TimeZone = record.TimeZone
	c.created = record.Created

	for k, v := range record.Doors {
		c.Doors[k] = v
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
		OID      string           `json:"OID"`
		Name     *types.Name      `json:"name,omitempty"`
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
		Created  time.Time        `json:"created"`
	}{
		OID:      c.OID,
		Name:     c.Name,
		DeviceID: c.DeviceID,
		Address:  c.IP,
		Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},
		TimeZone: c.TimeZone,
		Created:  c.created,
	}

	for k, v := range c.Doors {
		record.Doors[k] = v
	}

	return json.MarshalIndent(record, "", "  ")
}

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
	name := stringify(c.Name)
	deviceID := ""
	address := addr{}
	datetime := tinfo{}
	cards := cinfo{}
	events := einfo{}

	doors := map[uint8]string{1: "", 2: "", 3: "", 4: ""}

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
			case dt < DeviceOk:
				status = types.StatusOk

			case dt < DeviceUncertain:
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
				if delta <= WINDOW {
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
		object{OID: c.OID, Value: fmt.Sprintf("%v", status)},
		object{OID: c.OID + ".0.1", Value: created},
		object{OID: c.OID + ".1", Value: name},
		object{OID: c.OID + ".2", Value: deviceID},
		object{OID: c.OID + ".3", Value: address.address},
		object{OID: c.OID + ".3.1", Value: address.configured},
		object{OID: c.OID + ".3.2", Value: stringify(address.status)},
		object{OID: c.OID + ".4", Value: datetime.datetime},
		object{OID: c.OID + ".4.1", Value: datetime.system},
		object{OID: c.OID + ".4.2", Value: fmt.Sprintf("%v", datetime.status)},
		object{OID: c.OID + ".5", Value: cards.cards},
		object{OID: c.OID + ".5.1", Value: fmt.Sprintf("%v", cards.status)},
		object{OID: c.OID + ".6", Value: events.events},
		object{OID: c.OID + ".6.1", Value: fmt.Sprintf("%v", events.status)},
		object{OID: c.OID + ".7", Value: doors[1]},
		object{OID: c.OID + ".8", Value: doors[2]},
		object{OID: c.OID + ".9", Value: doors[3]},
		object{OID: c.OID + ".10", Value: doors[4]},
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

		re := regexp.MustCompile(`^door\[(.+?)\]\.(.*)$`)
		if match := re.FindStringSubmatch(f); match != nil && len(match) > 2 {
			oid := match[1]
			field := match[2]

			for k, v := range c.Doors {
				if v == oid && c.DeviceID != nil {
					if cached, ok := cache.cache[*c.DeviceID]; ok {
						if d, ok := cached.doors[k]; ok {
							switch field {
							case "mode":
								return d.mode
							case "delay":
								return d.delay
							case "control":
								return d.mode
							case "delay.dirty":
								return cached.dirty[fmt.Sprintf("door.%v.delay", k)]
							case "control.dirty":
								return cached.dirty[fmt.Sprintf("door.%v.control", k)]
							}
						}
					}
					break
				}
			}
		}
	}

	return nil
}

func (c *Controller) String() string {
	if c == nil {
		return ""
	}

	return fmt.Sprintf("%v (%v)", stringify(c.Name), stringify(c.DeviceID))
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

func (c *Controller) set(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateController(c, field, value)
	}

	if c != nil {
		switch oid {
		case c.OID + ".1":
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "name", stringify(c.Name), value)
				name := types.Name(value)
				c.Name = &name
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".1",
					Value: stringify(c.Name),
				})
			}

		case c.OID + ".2":
			if err := f("deviceID", value); err != nil {
				return nil, err
			} else if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if id, err := strconv.ParseUint(value, 10, 32); err == nil {
					c.log(auth, "update", c.OID, "device-id", stringify(c.DeviceID), value)
					cid := uint32(id)
					c.DeviceID = &cid
					c.unconfigured = false
					objects = append(objects, object{
						OID:   c.OID + ".2",
						Value: stringify(cid),
					})
				}
			} else if value == "" {
				c.log(auth, "update", c.OID, "device-id", stringify(c.DeviceID), value)
				c.DeviceID = nil
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".2",
					Value: "",
				})
			}

		case c.OID + ".3":
			if addr, err := core.ResolveAddr(value); err != nil {
				return nil, err
			} else if err := f("address", addr); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "address", stringify(c.IP), value)
				c.IP = addr
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".3",
					Value: fmt.Sprintf("%v", c.IP),
				})
			}

		case c.OID + ".4":
			if tz, err := types.Timezone(value); err != nil {
				return nil, err
			} else if err := f("timezone", tz); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "timezone", stringify(c.TimeZone), tz.String())
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

							objects = append(objects, object{
								OID:   c.OID + ".4",
								Value: dt.Format("2006-01-02 15:04 MST"),
							})
						}
					}
				}
			}

		case c.OID + ".7":
			if err := f("door[1]", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "door[1]", stringify(c.Doors[1]), value)
				c.Doors[1] = value
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".7",
					Value: fmt.Sprintf("%v", c.Doors[1]),
				})
			}

		case c.OID + ".8":
			if err := f("door[2]", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "door[2]", stringify(c.Doors[2]), value)
				c.Doors[2] = value
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".8",
					Value: fmt.Sprintf("%v", c.Doors[2]),
				})
			}

		case c.OID + ".9":
			if err := f("door[3]", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "door[3]", stringify(c.Doors[3]), value)
				c.Doors[3] = value
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".9",
					Value: fmt.Sprintf("%v", c.Doors[3]),
				})
			}

		case c.OID + ".10":
			if err := f("door[4]", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "door[4]", stringify(c.Doors[4]), value)
				c.Doors[4] = value
				c.unconfigured = false
				objects = append(objects, object{
					OID:   c.OID + ".10",
					Value: fmt.Sprintf("%v", c.Doors[4]),
				})
			}
		}

		if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
			if auth != nil {
				if err := auth.CanDeleteController(c); err != nil {
					return nil, err
				}
			}

			c.log(auth, "delete", c.OID, "device-id", "", "")
			now := time.Now()
			c.deleted = &now

			objects = append(objects, object{
				OID:   c.OID,
				Value: "deleted",
			})

			catalog.Delete(c.OID)
		}
	}

	return objects, nil
}

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			OID:      c.OID,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
			IP:       c.IP,
			TimeZone: c.TimeZone,
			Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},

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

func (c *Controller) log(auth auth.OpAuth, operation, OID, field, current, value string) {
	type info struct {
		OID        string `json:"OID"`
		Controller string `json:"controller"`
		Field      string `json:"field"`
		Current    string `json:"current"`
		Updated    string `json:"new"`
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
				OID:        OID,
				Controller: stringify(c.DeviceID),
				Field:      field,
				Current:    current,
				Updated:    value,
			},
		}

		trail.Write(record)
	}
}
