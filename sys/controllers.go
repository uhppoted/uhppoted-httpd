package system

import (
	"fmt"
	"math"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

// Container class for the static information pertaining to an access controller.
type Controller struct {
	ID       string           `json:"ID"`
	Created  time.Time        `json:"created"`
	Name     *types.Name      `json:"name"`
	DeviceID uint32           `json:"device-id"`
	IP       address          `json:"address"`
	Doors    map[uint8]string `json:"-"`
	TimeZone string           `json:"timezone"`
}

// Internal container class for the instantaneous state of an access controller. Used
// mostly for HTML templating.
type controller struct {
	Controller
	//	ID         string
	//	created    time.Time
	//	Name       *types.Name
	//	DeviceID   uint32
	IP         ip
	SystemTime datetime
	Cards      *records
	Events     *records
	//	Doors      map[uint8]string
	Status status
}

func merge(id uint32) controller {
	tz := time.Local

	c := controller{
		Controller: Controller{
			ID:       ID(id),
			DeviceID: id,
			Doors:    map[uint8]string{},
			Created:  time.Now(),
		},
	}

	for _, v := range sys.Controllers {
		if v.DeviceID == id {
			c.ID = v.ID
			c.Created = v.Created
			c.Name = v.Name
			c.IP = ip{
				IP: &v.IP,
			}

			for _, d := range sys.Doors {
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

	if cached, ok := sys.Local.cache[id]; ok {
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
