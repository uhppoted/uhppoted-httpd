package system

import (
	"fmt"
	"math"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

// Container class for the static information pertaining to an access controller.
type Controller struct {
	ID       string      `json:"ID"`
	Created  time.Time   `json:"created"`
	Name     *types.Name `json:"name"`
	DeviceID uint32      `json:"device-id"`
	IP       address     `json:"address"`
	TimeZone string      `json:"timezone"`
}

// Internal 'temporary' container class for the instantaneous state of an access controller.
// Used mostly for HTML templating.
type controller struct {
	Controller
	IP         ip
	SystemTime datetime
	Cards      *records
	Events     *records
	Doors      map[uint8]string
	Status     status
}

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			ID:       c.ID,
			Created:  c.Created,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
			IP:       c.IP,
			TimeZone: c.TimeZone,
		}

		return &replicant
	}

	return nil
}

func merge(id uint32) controller {
	tz := time.Local

	c := controller{
		Controller: Controller{
			ID:       ID(id),
			DeviceID: id,
			Created:  time.Now(),
		},
		Doors: map[uint8]string{},
	}

	for _, v := range sys.data.Tables.Controllers {
		if v.DeviceID == id {
			c.ID = v.ID
			c.Created = v.Created
			c.Name = v.Name
			c.IP = ip{
				IP: &v.IP,
			}

			for _, d := range sys.data.Tables.Doors {
				if d.DeviceID == c.DeviceID {
					c.Doors[d.Door] = d.Name
				}
			}

			if v.TimeZone != "" {
				if l, err := time.LoadLocation(v.TimeZone); err == nil {
					tz = l
				}
			}
		}
	}

	if cached, ok := sys.data.Tables.Local.cache[id]; ok {
		c.Cards = (*records)(cached.cards)
		c.Events = (*records)(cached.events)

		if cached.address != nil {
			switch {
			case c.IP.IP == nil:
				c.IP.Status = StatusUnknown

			case cached.address.Equal(c.IP.IP.IP):
				c.IP.Status = StatusOk

			default:
				c.IP.Status = StatusError
			}

			c.IP.IP = cached.address
		}

		if cached.datetime != nil {
			t := time.Time(*cached.datetime)
			T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
			delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

			if delta > WINDOW {
				c.SystemTime.Status = StatusError
			} else {
				c.SystemTime.Status = StatusOk
			}

			dt := types.DateTime(T)
			c.SystemTime.DateTime = &dt
			c.SystemTime.TimeZone = tz
		}

		switch dt := time.Now().Sub(cached.touched); {
		case dt < DeviceOk:
			c.Status = StatusOk
		case dt < DeviceUncertain:
			c.Status = StatusUncertain
		}
	}

	return c
}

func ID(id uint32) string {
	return fmt.Sprintf("L%d", id)
}
