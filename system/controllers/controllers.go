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
	"time"

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
	cc.Local.Init(cc.Controllers)

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
	record := c.Clone()

loop:
	for next := 1; ; next++ {
		oid := fmt.Sprintf("0.1.1.%v", next)
		for _, v := range cc.Controllers {
			if v.OID == oid {
				continue loop
			}
		}

		record.OID = oid
		break
	}

	record.Created = time.Now()

	cc.Controllers = append(cc.Controllers, record)

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
	for i, record := range cc.Controllers {
		if record.OID == c.OID {
			cc.Controllers = append(cc.Controllers[:i], cc.Controllers[i+1:]...)
			return &c, nil
		}
	}

	return nil, nil
}

func (cc *Controllers) Refresh() {
	cc.Local.refresh()
}

func (cc *Controllers) Clone() *Controllers {
	shadow := Controllers{
		file:        cc.file,
		Controllers: make([]*Controller, len(cc.Controllers)),
		Local:       &Local{},
	}

	for k, v := range cc.Controllers {
		shadow.Controllers[k] = v.Clone()
	}

	shadow.Local = cc.Local.clone()

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

func (cc *Controllers) Sync() error {
	return nil
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
			return &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid controller OID"),
				Detail: fmt.Errorf("Invalid controller OID (%v)", c.OID),
			}
		}

		if c.DeviceID != nil && *c.DeviceID != 0 {
			id := *c.DeviceID

			if cid, ok := devices[id]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate controller ID (%v)", id),
					Detail: fmt.Errorf("controller %v: duplicate device ID in records %v and %v", id, cid, c.OID),
				}
			}

			devices[id] = c.OID
		}
	}

	return nil
}

func scrub(cc *Controllers) error {
	return nil
}
