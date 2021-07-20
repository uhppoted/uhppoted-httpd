package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
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

	Status     status
	SystemTime datetime
	Cards      cards
	Events     *records
	Deleted    bool

	created time.Time
}

func (c *controller) Created() time.Time {
	if c != nil {
		return c.created
	}

	return time.Now()
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

	if (c.Name == nil || *c.Name != "") && (c.DeviceID == nil || *c.DeviceID == 0) {
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

func (c *Controller) AsView() interface{} {
	if c == nil {
		return nil
	}

	v := controller{
		OID:      c.OID,
		Name:     "",
		DeviceID: "",
		IP: ip{
			Configured: c.IP,
		},
		Doors: map[uint8]string{1: "", 2: "", 3: "", 4: ""},

		Status: StatusUnknown,
		SystemTime: datetime{
			Status: StatusUnknown,
		},
		Cards: cards{
			Status: StatusUnknown,
		},
		Deleted: c.deleted != nil,

		created: c.created,
	}

	if c.Name != nil {
		v.Name = fmt.Sprintf("%v", *c.Name)
	}

	if c.DeviceID != nil && *c.DeviceID != 0 {
		v.DeviceID = fmt.Sprintf("%v", *c.DeviceID)
	}

	if c.IP != nil {
		v.IP.Address = &(*c.IP)
	}

	for _, i := range []uint8{1, 2, 3, 4} {
		if d, ok := c.Doors[i]; ok {
			v.Doors[i] = d
		}
	}

	if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
		v.Status = StatusNew
	}

	if c.DeviceID == nil || *c.DeviceID == 0 {
		return &v
	}

	if cached, ok := cache.cache[*c.DeviceID]; ok {
		// ... set status field from cached value
		dt := time.Now().Sub(cached.touched)
		switch {
		case dt < DeviceOk:
			v.Status = StatusOk

		case dt < DeviceUncertain:
			v.Status = StatusUncertain
		}

		// ... set IP address field from cached value
		if cached.address != nil {
			v.IP.Address = &(*cached.address)

			switch {
			case c.IP == nil:
				v.IP.Status = StatusUnknown

			case cached.address.Equal(c.IP):
				v.IP.Status = StatusOk

			default:
				v.IP.Status = StatusError
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

			now := types.DateTime(time.Now().In(tz))
			t := time.Time(*cached.datetime)
			T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
			delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

			if delta > WINDOW {
				v.SystemTime.Status = StatusError
			} else {
				v.SystemTime.Status = StatusOk
			}

			dt := types.DateTime(T)
			v.SystemTime.DateTime = &dt
			v.SystemTime.Expected = &now
		}

		// ... set ACL field from cached value
		if cached.cards != nil {
			v.Cards.Records = records(*cached.cards)
			if cached.acl == StatusUnknown {
				v.Cards.Status = StatusUncertain
			} else {
				v.Cards.Status = cached.acl
			}
		}

		// ... set events field from cached value
		v.Events = (*records)(cached.events)
	}

	return &v
}

func (c *Controller) AsObjects() []interface{} {
	type addr struct {
		address    string
		configured string
		status     status
	}

	type tinfo struct {
		datetime string
		system   string
		status   status
	}

	created := c.created.Format("2006-01-02 15:04:05")
	status := StatusUnknown
	name := ""
	deviceID := ""
	address := addr{}
	datetime := tinfo{}
	cards := ""
	events := ""
	doors := map[uint8]string{1: "", 2: "", 3: "", 4: ""}

	if c.Name != nil {
		name = fmt.Sprintf("%v", *c.Name)
	}

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
				status = StatusOk

			case dt < DeviceUncertain:
				status = StatusUncertain
			}

			// ... set IP address field from cached value
			if cached.address != nil {
				address.address = fmt.Sprintf("%v", cached.address)
				switch {
				case c.IP == nil || (c.IP != nil && cached.address.Equal(c.IP)):
					address.status = StatusOk

				case c.IP != nil && !cached.address.Equal(c.IP):
					address.status = StatusError

				default:
					address.status = StatusUnknown
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
					datetime.status = StatusOk
				} else {
					datetime.status = StatusError
				}
			}

			// ... set ACL field from cached value
			if cached.cards != nil {
				cards = fmt.Sprintf("%d", *cached.cards)
				//			v.Cards.Records = records(*cached.cards)
				//			if cached.acl == StatusUnknown {
				//				v.Cards.Status = StatusUncertain
				//			} else {
				//				v.Cards.Status = cached.acl
				//			}
			}

			// ... set events field from cached value
			events = fmt.Sprintf("%v", (*records)(cached.events))
		}
	}

	if c.deleted != nil {
		status = StatusDeleted
	}

	objects := []interface{}{
		object{OID: c.OID, Value: fmt.Sprintf("%v", status)},
		object{OID: c.OID + ".0.1", Value: created},
		object{OID: c.OID + ".1", Value: name},
		object{OID: c.OID + ".2", Value: deviceID},
		object{OID: c.OID + ".3", Value: address.address},
		object{OID: c.OID + ".3.1", Value: address.configured},
		object{OID: c.OID + ".3.2", Value: fmt.Sprintf("%v", address.status)},
		object{OID: c.OID + ".4", Value: datetime.datetime},
		object{OID: c.OID + ".4.1", Value: datetime.system},
		object{OID: c.OID + ".4.2", Value: fmt.Sprintf("%v", datetime.status)},
		object{OID: c.OID + ".5", Value: cards},
		object{OID: c.OID + ".6", Value: events},
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

func (c *Controller) String() string {
	if c != nil {
		s := fmt.Sprintf("%v", c.OID)

		if c.Name != nil {
			s += fmt.Sprintf(" %v", *c.Name)
		}

		if c.DeviceID != nil {
			s += fmt.Sprintf(" %v", *c.DeviceID)
		}

		return s
	}

	return ""
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

func (c *Controller) set(oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	if c != nil {
		switch oid {
		case c.OID + ".1":
			name := types.Name(value)
			c.Name = &name
			objects = append(objects, object{
				OID:   c.OID + ".1",
				Value: fmt.Sprintf("%v", c.Name),
			})

		case c.OID + ".2":
			if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if id, err := strconv.ParseUint(value, 10, 32); err == nil {
					uid := uint32(id)
					c.DeviceID = &uid
					objects = append(objects, object{
						OID:   c.OID + ".2",
						Value: fmt.Sprintf("%v", uid),
					})
				}
			} else if value == "" {
				c.DeviceID = nil
				objects = append(objects, object{
					OID:   c.OID + ".2",
					Value: "",
				})
			}

		case c.OID + ".3":
			if addr, err := core.ResolveAddr(value); err != nil {
				return nil, err
			} else {
				c.IP = addr
				objects = append(objects, object{
					OID:   c.OID + ".3",
					Value: fmt.Sprintf("%v", c.IP),
				})
			}

		case c.OID + ".4":
			if tz, err := types.Timezone(value); err != nil {
				return nil, err
			} else {
				tzs := tz.String()
				c.TimeZone = &tzs

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

		case c.OID + ".7":
			c.Doors[1] = value
			objects = append(objects, object{
				OID:   c.OID + ".7",
				Value: fmt.Sprintf("%v", c.Doors[1]),
			})

		case c.OID + ".8":
			c.Doors[2] = value
			objects = append(objects, object{
				OID:   c.OID + ".8",
				Value: fmt.Sprintf("%v", c.Doors[2]),
			})

		case c.OID + ".9":
			c.Doors[3] = value
			objects = append(objects, object{
				OID:   c.OID + ".9",
				Value: fmt.Sprintf("%v", c.Doors[3]),
			})

		case c.OID + ".10":
			c.Doors[4] = value
			objects = append(objects, object{
				OID:   c.OID + ".10",
				Value: fmt.Sprintf("%v", c.Doors[4]),
			})
		}

		if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
			now := time.Now()
			c.deleted = &now

			objects = append(objects, object{
				OID:   c.OID,
				Value: "deleted",
			})
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
