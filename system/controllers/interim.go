package controllers

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

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
	Deleted    bool
}

func Consolidate(lan *LAN, controllers []*Controller) interface{} {
	devices := []Controller{}
	for _, v := range controllers {
		if v.IsValid() {
			devices = append(devices, *v)
		}
	}

loop:
	for k, _ := range lan.Cache {
		for _, c := range devices {
			if c.DeviceID != nil && *c.DeviceID == k {
				continue loop
			}
		}

		// ... include 'unconfigured' controllers
		id := k
		oid := catalog.Get(k)
		devices = append(devices, Controller{
			OID:      oid,
			DeviceID: &id,
			Created:  time.Now(),
		})
	}

	list := []controller{}
	for _, c := range devices {
		list = append(list, Merge(lan, c))
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].Created.Before(list[j].Created) })

	return list
}

func Merge(lan *LAN, c Controller) controller {
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
		Deleted: c.deleted != nil,
		Status:  StatusUnknown,
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

	if cached, ok := lan.Cache[*c.DeviceID]; ok {
		if cached.cards != nil {
			cc.Cards.Records = records(*cached.cards)
			if cached.acl == StatusUnknown {
				cc.Cards.Status = StatusUncertain
			} else {
				cc.Cards.Status = cached.acl
			}
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
			now := types.DateTime(time.Now().In(tz))
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
			cc.SystemTime.Expected = &now
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

func warn(err error) {
	log.Printf("ERROR %v", err)
}
