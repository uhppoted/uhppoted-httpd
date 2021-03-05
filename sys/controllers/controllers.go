package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	Controllers []*Controller `json:"controllers"`
	Local       *Local        `json:"local"`
}

var guard sync.Mutex

func NewControllers() Controllers {
	return Controllers{
		Controllers: []*Controller{},
		Local: &Local{
			Devices: map[uint32]types.Address{},
		},
	}
}

func (c *Controllers) Refresh() {
	c.Local.refresh()
}

func (c *Controllers) Clone() *Controllers {
	shadow := Controllers{
		Controllers: make([]*Controller, len(c.Controllers)),
		Local:       &Local{},
	}

	for k, v := range c.Controllers {
		shadow.Controllers[k] = v.Clone()
	}

	shadow.Local = c.Local.clone()

	return &shadow
}

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
