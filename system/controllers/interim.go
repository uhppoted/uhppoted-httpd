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

func merge(lan *LAN, c Controller) controller {
	cc := controller{
		OID:      c.OID,
		Name:     "",
		DeviceID: "",
		IP: ip{
			Configured: c.IP,
		},
		Doors: map[uint8]string{1: "", 2: "", 3: "", 4: ""},

		Status:     StatusUnknown,
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

	if cc.Name == "" && cc.DeviceID == "" {
		cc.Status = StatusNew
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

		switch dt := time.Now().Sub(cached.touched); {
		case dt < DeviceOk:
			cc.Status = StatusOk
		case dt < DeviceUncertain:
			cc.Status = StatusUncertain
		}
	}

	return cc
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
