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

type Interface struct {
	file      string        `json:"-"`
	retention time.Duration `json:"-"`
	Interface []*Controller `json:"controllers"`
	LAN       *LAN          `json:"LAN"`
}

var guard sync.Mutex

func NewInterface() Interface {
	return Interface{
		Interface: []*Controller{},
		LAN:       NewLAN(),
		retention: 6 * time.Hour,
	}
}

func (cc *Interface) Load(file string, retention time.Duration) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, cc)
	if err != nil {
		return err
	}

	cc.file = file
	cc.retention = retention
	cc.LAN.Init(cc.Interface)

	return nil
}

func (cc *Interface) Save() error {
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

	cleaned := Interface{
		file:      cc.file,
		retention: cc.retention,
		Interface: []*Controller{},
		LAN:       cc.LAN.clone(),
	}

	for _, v := range cc.Interface {
		if v.deleted == nil {
			cleaned.Interface = append(cleaned.Interface, v.clone())
		}
	}

	b, err := json.MarshalIndent(cleaned, "", "  ")
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

func (cc *Interface) Sweep() {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-cc.retention)
	for i, v := range cc.Interface {
		if v.deleted != nil && v.deleted.Before(cutoff) {
			cc.Interface = append(cc.Interface[:i], cc.Interface[i+1:]...)
		}
	}
}

func (cc *Interface) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	}
}

func (cc *Interface) Add(c Controller) (*Controller, error) {
	id := uint32(0)
	if c.DeviceID != nil {
		id = *c.DeviceID
	}

	record := c.clone()
	record.OID = catalog.Get(id)
	record.Created = time.Now()

	cc.Interface = append(cc.Interface, record)
	cc.LAN.add(*record)

	return record, nil
}

func (cc *Interface) Update(c Controller) (*Controller, error) {
	for _, record := range cc.Interface {
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

func (cc *Interface) Delete(c Controller) (*Controller, error) {
	for _, record := range cc.Interface {
		if record.OID == c.OID {
			now := time.Now()
			c.deleted = &now
			record.deleted = &now
			cc.LAN.delete(*record)
			return &c, nil
		}
	}

	return nil, nil
}

func (cc *Interface) Refresh() {
	devices := []uint32{}
	for _, record := range cc.Interface {
		if record.DeviceID != nil && *record.DeviceID != 0 && record.deleted == nil {
			devices = append(devices, *record.DeviceID)
		}
	}

	cc.LAN.refresh(devices)
}

func (cc *Interface) Clone() *Interface {
	shadow := Interface{
		file:      cc.file,
		retention: cc.retention,
		Interface: make([]*Controller, len(cc.Interface)),
		LAN:       &LAN{},
	}

	for k, v := range cc.Interface {
		shadow.Interface[k] = v.clone()
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
		if c.DeviceID != nil && *c.DeviceID != 0 && c.deleted == nil {
			device := config.Device{
				Name:     "",
				Address:  nil,
				Doors:    []string{"", "", "", ""},
				TimeZone: "",
				Rollover: 100000,
			}

			if c.Name != nil {
				device.Name = fmt.Sprintf("%v", c.Name)
			}

			if c.IP != nil {
				device.Address = (*net.UDPAddr)(c.IP)
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

func (cc *Interface) Sync() {
	for _, c := range cc.Interface {
		if c != nil {
			cc.LAN.synchTime(*c)
		}
	}
}

func (cc *Interface) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func validate(cc Interface) error {
	devices := map[uint32]string{}

	for _, c := range cc.Interface {
		if c.OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", c.OID)
		}

		if c.deleted != nil {
			continue
		}

		if c.DeviceID != nil && *c.DeviceID != 0 {
			id := *c.DeviceID

			if _, ok := devices[id]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", id)
			}

			devices[id] = c.OID
		}
	}

	for k, _ := range cc.LAN.cache {
		if oid, ok := devices[k]; ok {
			if oid != catalog.Find(k) {
				return fmt.Errorf("Duplicate controller ID (%v)", k)
			}
		}
	}

	return nil
}

func scrub(cc *Interface) error {
	return nil
}
