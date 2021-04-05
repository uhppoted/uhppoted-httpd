package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	OID      string           `json:"OID"`
	Name     *types.Name      `json:"name,omitempty"`
	DeviceID *uint32          `json:"device-id,omitempty"`
	IP       *types.Address   `json:"address,omitempty"`
	Doors    map[uint8]string `json:"doors"`
	TimeZone *string          `json:"timezone,omitempty"`

	Status     status   `json:"status"`
	SystemTime datetime `json:"systime"`
	Cards      cards    `json:"cards,omitempty"`
	Events     *records `json:"events,omitempty"`

	created      time.Time
	deleted      *time.Time
	touched      time.Time
	unconfigured bool
}

func (c *Controller) deserialize(bytes []byte) error {
	record := struct {
		OID      string           `json:"OID"`
		Name     *types.Name      `json:"name,omitempty"`
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *types.Address   `json:"address,omitempty"`
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
	c.Status = StatusUnknown
	c.SystemTime = datetime{
		Status: StatusUnknown,
	}
	c.Cards = cards{
		Status: StatusUnknown,
	}

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
		Address  *types.Address   `json:"address,omitempty"`
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

func (c *Controller) AsView(lan *LAN) interface{} {
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

	if cached, ok := lan.cache[*c.DeviceID]; ok {
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

		dt := time.Now().Sub(cached.touched)
		switch {
		case dt < DeviceOk:
			v.Status = StatusOk

		case dt < DeviceUncertain:
			v.Status = StatusUncertain
		}
	}

	return &v
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

func (c *Controller) clone() *Controller {
	if c != nil {
		replicant := Controller{
			OID:      c.OID,
			Name:     c.Name.Copy(),
			DeviceID: c.DeviceID,
			IP:       c.IP,
			TimeZone: c.TimeZone,
			Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},

			Status:     c.Status,
			SystemTime: c.SystemTime,
			Cards: cards{
				Records: c.Cards.Records,
				Status:  c.Cards.Status,
			},
			Events: c.Events,

			created:      c.created,
			deleted:      c.deleted,
			touched:      c.touched,
			unconfigured: c.unconfigured,
		}

		for k, v := range c.Doors {
			replicant.Doors[k] = v
		}

		return &replicant
	}

	return nil
}
