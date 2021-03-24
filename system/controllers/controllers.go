package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	file        string        `json:"-"`
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

var guard sync.Mutex

func NewControllers() Controllers {
	return Controllers{
		Controllers: []*Controller{},
		LAN:         NewLAN(),
	}
}

func (cc *Controllers) Load(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, cc)
	if err != nil {
		return err
	}

	cc.file = file
	cc.LAN.Init(cc.Controllers)

	return nil
}

func (cc *Controllers) Save() error {
	if cc == nil {
		return nil
	}

	if err := validate(*cc); err != nil {
		return err
	}

	if err := scrub(cc); err != nil {
		return err
	}

	if cc.file == "" {
		return nil
	}

	b, err := json.MarshalIndent(cc, "", "  ")
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

	if err := os.MkdirAll(filepath.Dir(cc.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), cc.file)
}

func (cc *Controllers) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	}
}

func (cc *Controllers) Add(c Controller) (*Controller, error) {
	id := uint32(0)
	if c.DeviceID != nil {
		id = *c.DeviceID
	}

	record := c.clone()
	record.OID = catalog.Get(id)
	record.Created = time.Now()

	cc.Controllers = append(cc.Controllers, record)
	cc.LAN.add(*record)

	return record, nil
}

func (cc *Controllers) Update(c Controller) (*Controller, error) {
	for _, record := range cc.Controllers {
		if record.OID == c.OID {
			if c.Name != nil {
				record.Name = c.Name
			}

			if c.DeviceID != nil && *c.DeviceID != 0 {
				id := *c.DeviceID
				record.DeviceID = &id
			}

			if c.IP != nil {
				record.IP = c.IP.Clone()
			}

			if c.TimeZone != nil {
				tz := *c.TimeZone
				record.TimeZone = &tz
			}

			if c.Doors != nil {
				for k, v := range c.Doors {
					record.Doors[k] = v
				}
			}

			return record, nil
		}
	}

	return nil, fmt.Errorf("Invalid controller OID '%v'", c.OID)
}

func (cc *Controllers) Delete(c Controller) (*Controller, error) {
	for _, record := range cc.Controllers {
		if record.OID == c.OID {
			c.deleted = true
			record.deleted = true
			cc.LAN.delete(*record)
			return &c, nil
		}
	}

	return nil, nil
}

func (cc *Controllers) Refresh() {
	devices := []uint32{}
	for _, record := range cc.Controllers {
		if record.DeviceID != nil && *record.DeviceID != 0 {
			devices = append(devices, *record.DeviceID)
		}
	}

	cc.LAN.refresh(devices)
}

func (cc *Controllers) Clone() *Controllers {
	shadow := Controllers{
		file:        cc.file,
		Controllers: make([]*Controller, len(cc.Controllers)),
		LAN:         &LAN{},
	}

	for k, v := range cc.Controllers {
		shadow.Controllers[k] = v.clone()
	}

	shadow.LAN = cc.LAN.clone()

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

func (cc *Controllers) Sync() {
	for _, c := range cc.Controllers {
		if c != nil {
			cc.LAN.synchTime(*c)
		}
	}
}

func (cc *Controllers) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func validate(cc Controllers) error {
	devices := map[uint32]string{}

	for _, c := range cc.Controllers {
		if c.OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", c.OID)
		}

		if c.DeviceID != nil && *c.DeviceID != 0 {
			id := *c.DeviceID

			if _, ok := devices[id]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", id)
			}

			devices[id] = c.OID
		}
	}

	return nil
}

func scrub(cc *Controllers) error {
	return nil
}
