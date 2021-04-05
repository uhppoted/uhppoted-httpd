package controllers

import (
	"fmt"
	"log"
	"time"
)

// Merges Controller static configuration with current controller state information into a struct usable
// by Javascript/HTML templating.
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

func merge(lan *LAN, c *Controller) controller {
	touched := c.touched

	if c.DeviceID != nil && *c.DeviceID != 0 {
		if cached, ok := lan.cache[*c.DeviceID]; ok {
			touched = cached.touched
		}
	}

	dt := time.Now().Sub(touched)
	switch {
	case (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0):
		c.Status = StatusNew

	case dt < DeviceOk:
		c.Status = StatusOk

	case dt < DeviceUncertain:
		c.Status = StatusUncertain

	default:
		c.Status = StatusUnknown

	}

	cc := controller{
		OID:      c.OID,
		Name:     "",
		DeviceID: "",
		IP: ip{
			Configured: c.IP,
		},
		Doors: map[uint8]string{1: "", 2: "", 3: "", 4: ""},

		Status:     c.Status,
		SystemTime: c.SystemTime,
		Cards: cards{
			Records: c.Cards.Records,
			Status:  c.Cards.Status,
		},
		Events:  c.Events,
		Deleted: c.deleted != nil,

		created: c.created,
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

	if c.DeviceID == nil || *c.DeviceID == 0 {
		return cc
	}

	if cached, ok := lan.cache[*c.DeviceID]; ok {
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
	}

	return cc
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
