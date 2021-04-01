package controllers

import (
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	OID          string           `json:"OID"`
	Created      time.Time        `json:"created"`
	Name         *types.Name      `json:"name,omitempty"`
	DeviceID     *uint32          `json:"device-id,omitempty"`
	IP           *types.Address   `json:"address,omitempty"`
	Doors        map[uint8]string `json:"doors"`
	TimeZone     *string          `json:"timezone,omitempty"`
	deleted      *time.Time
	unconfigured bool
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
			OID:          c.OID,
			Created:      c.Created,
			Name:         c.Name.Copy(),
			DeviceID:     c.DeviceID,
			IP:           c.IP,
			TimeZone:     c.TimeZone,
			Doors:        map[uint8]string{1: "", 2: "", 3: "", 4: ""},
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
