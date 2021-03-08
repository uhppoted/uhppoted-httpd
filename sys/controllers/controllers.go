package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	file        string        `json:"-"`
	Controllers []*Controller `json:"controllers"`
	Local       *Local        `json:"local"`
}

var guard sync.Mutex

func NewControllers() Controllers {
	return Controllers{
		Controllers: []*Controller{},
		Local:       NewLocal(),
	}
}

func (c *Controllers) Load(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, c)
	if err != nil {
		return err
	}

	c.file = file
	c.Local.Init(c.Controllers)

	return nil
}

func (c *Controllers) Save() error {
	if c == nil {
		return nil
	}

	if err := validate(*c); err != nil {
		return err
	}

	if err := scrub(c); err != nil {
		return err
	}

	if c.file == "" {
		return nil
	}

	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted-controllers.json")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), c.file)
}

func (c *Controllers) Print() {
	if b, err := json.MarshalIndent(c, "", "  "); err == nil {
		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	}
}

func (c *Controllers) Refresh() {
	c.Local.refresh()
}

func (c *Controllers) Clone() *Controllers {
	shadow := Controllers{
		file:        c.file,
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

func (c *Controllers) Validate() error {
	if c != nil {
		return validate(*c)
	}

	return nil
}

func validate(c Controllers) error {
	devices := map[uint32]string{}

	for _, r := range c.Controllers {
		if r.OID == "" {
			return &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid controller OID"),
				Detail: fmt.Errorf("Invalid controller OID (%v)", r.OID),
			}
		}

		if r.DeviceID != nil && *r.DeviceID != 0 {
			id := *r.DeviceID

			if rid, ok := devices[id]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate controller ID (%v)", id),
					Detail: fmt.Errorf("controller %v: duplicate device ID in records %v and %v", id, rid, r.OID),
				}
			}

			devices[id] = r.OID
		}
	}

	return nil
}

func scrub(c *Controllers) error {
	return nil
}
