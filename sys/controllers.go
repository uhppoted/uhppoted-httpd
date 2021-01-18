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

func merge(c Controller) controller {
	cc := controller{
		Controller: c,
		IP: ip{
			IP: &c.IP,
		},
		Doors: map[uint8]string{},
	}

	for _, d := range sys.data.Tables.Doors {
		if d.DeviceID == cc.DeviceID {
			cc.Doors[d.Door] = d.Name
		}
	}

	tz := time.Local
	if c.TimeZone != "" {
		if l, err := time.LoadLocation(c.TimeZone); err == nil {
			tz = l
		}
	}

	if cached, ok := sys.data.Tables.Local.cache[c.DeviceID]; ok {
		cc.Cards = (*records)(cached.cards)
		cc.Events = (*records)(cached.events)

		if cached.address != nil {
			switch {
			case cc.IP.IP == nil:
				cc.IP.Status = StatusUnknown

			case cached.address.Equal(cc.IP.IP.IP):
				cc.IP.Status = StatusOk

			default:
				cc.IP.Status = StatusError
			}

			cc.IP.IP = cached.address
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

func ID(id uint32) string {
	return fmt.Sprintf("L%d", id)
}
