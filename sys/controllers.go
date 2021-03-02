package system

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/types"
)

// Container class for the static information pertaining to an access controller.
type Controller struct {
	ID       string           `json:"-"`
	OID      string           `json:"OID"`
	Created  time.Time        `json:"created"`
	Name     *types.Name      `json:"name"`
	DeviceID *uint32          `json:"device-id"`
	IP       *address         `json:"address"`
	Doors    map[uint8]string `json:"doors"`
	TimeZone *string          `json:"timezone"`
}

// Merges Controller static configuration with current controller state information into a struct usable
// by Javascript/HTML templating.
type controller struct {
	ID       string
	OID      string
	Name     string
	DeviceID string
	IP       ip
	Doors    map[uint8]string

	Created time.Time

	SystemTime datetime
	Cards      cards
	Events     *records
	Status     status
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

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			ID:       c.ID,
			OID:      c.OID,
			Created:  c.Created,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
			IP:       c.IP,
			TimeZone: c.TimeZone,
			Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},
		}

		for k, v := range c.Doors {
			replicant.Doors[k] = v
		}

		return &replicant
	}

	return nil
}

func merge(c Controller) controller {
	cc := controller{
		ID:       ID(c),
		Name:     "",
		OID:      c.OID,
		DeviceID: "",
		IP: ip{
			Configured: c.IP,
		},
		Cards: cards{
			Status: StatusUnknown,
		},
		Doors: map[uint8]string{1: "", 2: "", 3: "", 4: ""},

		Created: c.Created,
	}

	if c.Name != nil {
		cc.Name = fmt.Sprintf("%v", *c.Name)
	}

	if c.DeviceID != nil {
		cc.DeviceID = fmt.Sprintf("%v", *c.DeviceID)
	}

	if c.IP != nil {
		cc.IP.Address = &(*c.IP)
	}

	for _, i := range []uint8{1, 2, 3, 4} {
		if d, ok := c.Doors[i]; ok {
			cc.Doors[i] = d
		}
	}

	tz := time.Local
	if c.TimeZone != nil {
		if l, err := timezone(*c.TimeZone); err != nil {
			warn(err)
		} else {
			tz = l
		}
	}

	if c.DeviceID == nil || *c.DeviceID == 0 {
		return cc
	}

	if cached, ok := sys.data.Tables.Local.cache[*c.DeviceID]; ok {
		if cached.cards != nil {
			cc.Cards.Records = records(*cached.cards)
			cc.Cards.Status = StatusOk
		}

		cc.Events = (*records)(cached.events)

		if cached.address != nil {
			cc.IP.Address = &(*cached.address)

			switch {
			case c.IP == nil:
				cc.IP.Status = StatusUnknown

			case cached.address.Equal(c.IP):
				cc.IP.Status = StatusOk

			default:
				cc.IP.Status = StatusError
			}
		}

		if cached.datetime != nil {
			t := time.Time(*cached.datetime)
			T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
			delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

			if delta > WINDOW {
				cc.SystemTime.Status = StatusError
			} else {
				cc.SystemTime.Status = StatusOk
			}

			dt := types.DateTime(T)
			cc.SystemTime.DateTime = &dt
			cc.SystemTime.TimeZone = tz
		}

		switch dt := time.Now().Sub(cached.touched); {
		case dt < DeviceOk:
			cc.Status = StatusOk
		case dt < DeviceUncertain:
			cc.Status = StatusUncertain
		}
	}

	return cc
}

func ID(c Controller) string {
	if c.ID != "" {
		return c.ID
	}

	if c.OID != "" {
		return fmt.Sprintf("O%s", strings.ReplaceAll(c.OID, ".", ""))
	}

	uuid := strings.ReplaceAll(uuid.New().String(), "-", "")
	if uuid == "" {
		uuid = fmt.Sprintf("%d", time.Now().Unix())
	}

	return "U" + strings.ReplaceAll(uuid, "-", "")
}
