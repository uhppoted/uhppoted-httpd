package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type ControllerSet struct {
	file        string        `json:"-"`
	retention   time.Duration `json:"-"`
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

var guard sync.Mutex

func NewControllerSet() ControllerSet {
	return ControllerSet{
		Controllers: []*Controller{},
		LAN:         &LAN{},
		retention:   6 * time.Hour,
	}
}

func (cc *ControllerSet) Load(file string, retention time.Duration) error {
	type controller struct {
		OID      string           `json:"OID"`
		Name     *types.Name      `json:"name,omitempty"`
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *types.Address   `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
		Created  time.Time        `json:"created"`
	}

	blob := struct {
		Controllers []controller `json:"controllers"`
		LAN         *LAN         `json:"LAN"`
	}{
		Controllers: []controller{},
		LAN:         &LAN{},
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	cc.file = file
	cc.retention = retention
	cc.Controllers = []*Controller{}
	cc.LAN = &LAN{
		BindAddress:      blob.LAN.BindAddress,
		BroadcastAddress: blob.LAN.BroadcastAddress,
		ListenAddress:    blob.LAN.ListenAddress,
		Debug:            blob.LAN.Debug,
		cache:            map[uint32]device{},
	}

	for _, c := range blob.Controllers {
		replicant := Controller{
			OID:      c.OID,
			Name:     c.Name,
			DeviceID: c.DeviceID,
			IP:       c.Address,
			Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},
			TimeZone: c.TimeZone,

			SystemTime: datetime{
				Status: StatusUnknown,
			},
			Cards: cards{
				Status: StatusUnknown,
			},

			created: c.Created,
		}

		for k, v := range c.Doors {
			replicant.Doors[k] = v
		}

		cc.Controllers = append(cc.Controllers, &replicant)
	}

	for _, v := range cc.Controllers {
		if v.DeviceID != nil && *v.DeviceID != 0 {
			catalog.Put(*v.DeviceID, v.OID)
		}
	}

	return nil
}

func (cc *ControllerSet) Save() error {
	type controller struct {
		OID      string           `json:"OID"`
		Name     *types.Name      `json:"name,omitempty"`
		DeviceID *uint32          `json:"device-id,omitempty"`
		Address  *types.Address   `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
		Created  time.Time        `json:"created"`
	}

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

	cleaned := struct {
		Controllers []controller `json:"controllers"`
		LAN         *LAN         `json:"LAN"`
	}{
		Controllers: []controller{},
		LAN:         cc.LAN.clone(),
	}

	for _, c := range cc.Controllers {
		if c.IsSaveable() {
			replicant := controller{
				OID:      c.OID,
				Name:     c.Name,
				DeviceID: c.DeviceID,
				Address:  c.IP,
				Doors:    map[uint8]string{1: "", 2: "", 3: "", 4: ""},
				TimeZone: c.TimeZone,
				Created:  c.created,
			}

			for k, v := range c.Doors {
				replicant.Doors[k] = v
			}

			cleaned.Controllers = append(cleaned.Controllers, replicant)
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

func (cc *ControllerSet) Sweep() {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-cc.retention)
	for i, v := range cc.Controllers {
		if v.deleted != nil && v.deleted.Before(cutoff) {
			cc.Controllers = append(cc.Controllers[:i], cc.Controllers[i+1:]...)
		}
	}
}

func (cc *ControllerSet) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	}
}

func (cc *ControllerSet) Add(c Controller) (*Controller, error) {
	id := uint32(0)
	if c.DeviceID != nil {
		id = *c.DeviceID
	}

	record := c.clone()
	record.OID = catalog.Get(id)
	record.created = time.Now()

	cc.Controllers = append(cc.Controllers, record)

	return record, nil
}

func (cc *ControllerSet) Update(c Controller) (*Controller, error) {
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

			record.unconfigured = false

			return record, nil
		}
	}

	return nil, fmt.Errorf("Invalid controller OID '%v'", c.OID)
}

func (cc *ControllerSet) Delete(c Controller) (*Controller, error) {
	for _, record := range cc.Controllers {
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

func (cc *ControllerSet) Consolidate() interface{} {
	list := []controller{}
	for _, c := range cc.Controllers {
		if c.IsValid() {
			list = append(list, merge(cc.LAN, *c))
		}
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].created.Before(list[j].created) })

	return list
}

func (cc *ControllerSet) Merge(c Controller) controller {
	return merge(cc.LAN, c)
}

func (cc *ControllerSet) Refresh() {
	cc.LAN.refresh(cc.Controllers)

	// ... add 'found' controllers to list
loop:
	for k, _ := range cc.LAN.cache {
		for _, c := range cc.Controllers {
			if c.DeviceID != nil && *c.DeviceID == k && c.deleted == nil {
				continue loop
			}
		}

		id := k
		oid := catalog.Get(k)

		cc.Controllers = append(cc.Controllers, &Controller{
			OID:          oid,
			DeviceID:     &id,
			created:      time.Now(),
			unconfigured: true,
		})
	}

	// ... update from cache

	for _, c := range cc.Controllers {
		if c.DeviceID != nil && *c.DeviceID != 0 {
			if cached, ok := cc.LAN.cache[*c.DeviceID]; ok {
				if cached.datetime != nil {
					tz := time.Local
					if c.TimeZone != nil {
						if l, err := timezone(*c.TimeZone); err != nil {
							warn(err)
						} else {
							tz = l
						}
					}

					now := types.DateTime(time.Now().In(tz))
					t := time.Time(*cached.datetime)
					T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
					delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

					if delta > WINDOW {
						c.SystemTime.Status = StatusError
					} else {
						c.SystemTime.Status = StatusOk
					}

					dt := types.DateTime(T)
					c.SystemTime.DateTime = &dt
					c.SystemTime.Expected = &now
				}

				if cached.cards != nil {
					c.Cards.Records = records(*cached.cards)
					if cached.acl == StatusUnknown {
						c.Cards.Status = StatusUncertain
					} else {
						c.Cards.Status = cached.acl
					}
				}

				c.Events = (*records)(cached.events)
			}
		}
	}
}

func (cc *ControllerSet) Clone() *ControllerSet {
	shadow := ControllerSet{
		file:        cc.file,
		retention:   cc.retention,
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

func (cc *ControllerSet) Sync() {
	cc.LAN.synchTime(cc.Controllers)
}

func (cc *ControllerSet) Compare(permissions acl.ACL) error {
	return cc.LAN.compareACL(cc.Controllers, permissions)
}

func (cc *ControllerSet) UpdateACL(acl acl.ACL) {
	cc.LAN.updateACL(cc.Controllers, acl)
}

func (cc *ControllerSet) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func validate(cc ControllerSet) error {
	devices := map[uint32]string{}

	for _, c := range cc.Controllers {
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

	return nil
}

func scrub(cc *ControllerSet) error {
	return nil
}
