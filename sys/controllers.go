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
	OID      string           `json:"OID"`
	Created  time.Time        `json:"created"`
	Name     *types.Name      `json:"name"`
	DeviceID *uint32          `json:"device-id"`
	IP       *address         `json:"address"`
	Doors    map[uint8]string `json:"doors"`
	TimeZone string           `json:"timezone"`
}

// Internal 'temporary' container class for the instantaneous state of an access controller.
// Used mostly for HTML templating.
type controller struct {
	ID string
	Controller
	IP         ip
	SystemTime datetime
	Cards      *records
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
			OID:      c.OID,
			Created:  c.Created,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
			IP:       c.IP,
			TimeZone: c.TimeZone,
			Doors:    map[uint8]string{},
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
		ID:         ID(c),
		Controller: c,
		IP:         ip{},
	}

	if c.IP != nil {
		cc.IP.IP = &(*c.IP)
	}

	//	for _, d := range sys.data.Tables.Doors {
	//		if cc.DeviceID != nil && *cc.DeviceID != 0 && d.DeviceID == *cc.DeviceID {
	//			cc.Doors[d.Door] = d.ID
	//		}
	//	}

	tz := time.Local
	if c.TimeZone != "" {
		if l, err := time.LoadLocation(c.TimeZone); err == nil {
			tz = l
		}
	}

	if c.DeviceID == nil || *c.DeviceID == 0 {
		return cc
	}

	if cached, ok := sys.data.Tables.Local.cache[*c.DeviceID]; ok {
		cc.Cards = (*records)(cached.cards)
		cc.Events = (*records)(cached.events)

		if cached.address != nil {
			cc.IP.IP = &(*cached.address)

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
	if c.OID != "" {
		return fmt.Sprintf("O%s", strings.ReplaceAll(c.OID, ".", ""))
	}

	uuid := strings.ReplaceAll(uuid.New().String(), "-", "")
	if uuid == "" {
		uuid = fmt.Sprintf("%d", time.Now().Unix())
	}

	return "U" + strings.ReplaceAll(uuid, "-", "")
}
