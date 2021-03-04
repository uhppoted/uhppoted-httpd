package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controller struct {
	ID       string           `json:"-"` // TODO REMOVE
	OID      string           `json:"OID"`
	Created  time.Time        `json:"created"`
	Name     *types.Name      `json:"name"`
	DeviceID *uint32          `json:"device-id"`
	IP       *types.Address   `json:"address"`
	Doors    map[uint8]string `json:"doors"`
	TimeZone *string          `json:"timezone"`
}

var guard sync.Mutex

func Export(file string, controllers []*Controller, doors map[string]types.Door) error {
	guard.Lock()

	defer guard.Unlock()

	conf := config.NewConfig()
	if err := conf.Load(file); err != nil {
		return err
	}

	devices := config.DeviceMap{}
	for _, c := range controllers {
		if c.DeviceID != nil {
			device := config.Device{
				Name:     "",
				Address:  (*net.UDPAddr)(c.IP),
				Doors:    []string{"", "", "", ""},
				TimeZone: "",
				Rollover: 100000,
			}

			if c.Name != nil {
				device.Name = fmt.Sprintf("%v", c.Name)
			}

			if c.TimeZone != nil {
				device.TimeZone = *c.TimeZone
			}

			if d, ok := doors[c.Doors[1]]; ok {
				device.Doors[0] = d.Name
			}

			if d, ok := doors[c.Doors[2]]; ok {
				device.Doors[1] = d.Name
			}

			if d, ok := doors[c.Doors[3]]; ok {
				device.Doors[2] = d.Name
			}

			if d, ok := doors[c.Doors[4]]; ok {
				device.Doors[3] = d.Name
			}

			devices[*c.DeviceID] = &device
		}
	}

	conf.Devices = devices

	var b bytes.Buffer
	conf.Write(&b)

	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted.conf_")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b.Bytes()); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), file)
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

func (c *Controller) Clone() *Controller {
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
