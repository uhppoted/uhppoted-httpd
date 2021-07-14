package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

type ControllerSet struct {
	file        string        `json:"-"`
	retention   time.Duration `json:"-"`
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

type sortable interface {
	Created() time.Time
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
		Address  *core.Address    `json:"address,omitempty"`
		Doors    map[uint8]string `json:"doors"`
		TimeZone *string          `json:"timezone,omitempty"`
		Created  time.Time        `json:"created"`
	}

	blob := struct {
		Controllers []json.RawMessage `json:"controllers"`
		LAN         *LAN              `json:"LAN"`
	}{
		Controllers: []json.RawMessage{},
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
		OID:              blob.LAN.OID,
		Name:             blob.LAN.Name,
		BindAddress:      blob.LAN.BindAddress,
		BroadcastAddress: blob.LAN.BroadcastAddress,
		ListenAddress:    blob.LAN.ListenAddress,
		Debug:            blob.LAN.Debug,
	}

	for _, v := range blob.Controllers {
		var c Controller
		if err := c.deserialize(v); err == nil {
			cc.Controllers = append(cc.Controllers, &c)
		}
	}

	catalog.PutInterface(cc.LAN.OID)
	for _, v := range cc.Controllers {
		if v.DeviceID != nil && *v.DeviceID != 0 {
			catalog.PutController(*v.DeviceID, v.OID)
		}
	}

	return nil
}

func (cc *ControllerSet) Save() error {
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

	serializable := struct {
		Controllers []json.RawMessage `json:"controllers"`
		LAN         *LAN              `json:"LAN"`
	}{
		Controllers: []json.RawMessage{},
		LAN:         cc.LAN.clone(),
	}

	for _, c := range cc.Controllers {
		if record, err := c.serialize(); err == nil && record != nil {
			serializable.Controllers = append(serializable.Controllers, record)
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
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

func (cc *ControllerSet) UpdateByOID(oid string, value string) ([]interface{}, []interface{}, error) {
	if cc == nil {
		return nil, nil, nil
	}

	// ... interface
	if cc.LAN != nil && strings.HasPrefix(oid, cc.LAN.OID) {
		return cc.LAN.set(oid, value)
	}

	// ... controllers
	for _, c := range cc.Controllers {
		if c != nil && strings.HasPrefix(oid, c.OID) {
			return c.set(oid, value)
		}
	}

	updated := []interface{}{}
	added := []interface{}{}

	if oid == "<new>" {
		if c, err := cc.Add(Controller{}); err != nil {
			return nil, nil, err
		} else if c == nil {
			return nil, nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			type object struct {
				OID   string `json:"OID"`
				Value string `json:"value"`
			}

			added = append(added, object{
				OID:   c.OID,
				Value: "{}",
			})
		}
	}

	return updated, added, nil
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

func (cc *ControllerSet) AsView() interface{} {
	lan := struct {
		OID              string `json:"OID"`
		Type             string `json:"type"`
		Name             string `json:"name"`
		BindAddress      string `json:"bind-address"`
		BroadcastAddress string `json:"broadcast-address"`
		ListenAddress    string `json:"listen-address"`
	}{
		OID:              cc.LAN.OID,
		Type:             "LAN",
		Name:             cc.LAN.Name,
		BindAddress:      fmt.Sprintf("%v", cc.LAN.BindAddress),
		BroadcastAddress: fmt.Sprintf("%v", cc.LAN.BroadcastAddress),
		ListenAddress:    fmt.Sprintf("%v", cc.LAN.ListenAddress),
	}

	list := []interface{}{}
	for _, c := range cc.Controllers {
		if c.IsValid() {
			if record := c.AsView(); record != nil {
				list = append(list, record)
			}
		}
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].(sortable).Created().Before(list[j].(sortable).Created())
	})

	return struct {
		Interface   interface{} `json:"interface"`
		Controllers interface{} `json:"controllers"`
	}{
		Interface:   lan,
		Controllers: list,
	}
}

func (cc *ControllerSet) Refresh() {
	cc.LAN.refresh(cc.Controllers)

	// ... add 'found' controllers to list
loop:
	for k, _ := range cache.cache {
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

func warn(err error) {
	log.Printf("ERROR %v", err)
}
