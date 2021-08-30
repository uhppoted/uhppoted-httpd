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
	"strings"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

type ControllerSet struct {
	file        string        `json:"-"`
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

type object catalog.Object

var guard sync.Mutex
var trail audit.Trail

var windows = struct {
	deviceOk        time.Duration
	deviceUncertain time.Duration
	systime         time.Duration
	cacheExpiry     time.Duration
}{
	deviceOk:        60 * time.Second,
	deviceUncertain: 300 * time.Second,
	systime:         300 * time.Second,
	cacheExpiry:     120 * time.Second,
}

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func SetWindows(ok, uncertain, systime, cacheExpiry time.Duration) {
	windows.deviceOk = ok
	windows.deviceUncertain = uncertain
	windows.systime = systime
	windows.cacheExpiry = cacheExpiry
}

func NewControllerSet() ControllerSet {
	return ControllerSet{
		Controllers: []*Controller{},
		LAN: &LAN{
			OID:    "0.1.1.1.1",
			status: types.StatusOk,
		},
	}
}

func (cc *ControllerSet) AsObjects() []interface{} {
	objects := []interface{}{}
	lan := cc.LAN.AsObjects()

	objects = append(objects, lan...)

	for _, c := range cc.Controllers {
		if c.IsValid() {
			if l := c.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (cc *ControllerSet) Load(file string) error {
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
		LAN: &LAN{
			status: types.StatusOk,
		},
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
	cc.Controllers = []*Controller{}
	cc.LAN = &LAN{
		OID:              blob.LAN.OID,
		Name:             blob.LAN.Name,
		BindAddress:      blob.LAN.BindAddress,
		BroadcastAddress: blob.LAN.BroadcastAddress,
		ListenAddress:    blob.LAN.ListenAddress,
		Debug:            blob.LAN.Debug,
		status:           types.StatusOk,
	}

	for _, v := range blob.Controllers {
		var c Controller
		if err := c.deserialize(v); err == nil {
			cc.Controllers = append(cc.Controllers, &c)
		}
	}

	catalog.PutInterface(cc.LAN.OID)
	for _, c := range cc.Controllers {
		if c.DeviceID != nil && *c.DeviceID != 0 {
			catalog.PutController(*c.DeviceID, c.OID)
		}

		for _, d := range []uint8{1, 2, 3, 4} {
			if oid, ok := c.Doors[d]; ok && oid != "" {
				catalog.PutV(oid+catalog.DoorControllerOID, c.OID, false)
				catalog.PutV(oid+catalog.DoorControllerCreated, c.created, false)
				catalog.PutV(oid+catalog.DoorControllerName, c.Name, false)
				catalog.PutV(oid+catalog.DoorControllerID, c.DeviceID, false)
				catalog.PutV(oid+catalog.DoorControllerDoor, d, false)
			}
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

func (cc *ControllerSet) Sweep(retention time.Duration) {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-retention)
	for i, v := range cc.Controllers {
		if v.deleted != nil && v.deleted.Before(cutoff) {
			cc.Controllers = append(cc.Controllers[:i], cc.Controllers[i+1:]...)
		}
	}
}

func (cc *ControllerSet) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("----------------- CONTROLLERS\n%s\n", string(b))
	}
}

func (cc *ControllerSet) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if cc == nil {
		return nil, nil
	}

	// ... interface
	if cc.LAN != nil && strings.HasPrefix(oid, cc.LAN.OID+".") {
		return cc.LAN.set(auth, oid, value)
	}

	// ... controllers
	for _, c := range cc.Controllers {
		if c != nil && strings.HasPrefix(oid, c.OID+".") {
			return c.set(auth, oid, value)
		}
	}

	objects := []interface{}{}

	if oid == "<new>" {
		if c, err := cc.add(auth, Controller{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			c.log(auth, "add", c.OID, "controller", "", "")
			objects = append(objects, object{
				OID:   c.OID,
				Value: "new",
			})
		}
	}

	return objects, nil
}

func (cc *ControllerSet) add(auth auth.OpAuth, c Controller) (*Controller, error) {
	id := uint32(0)
	if c.DeviceID != nil {
		id = *c.DeviceID
	}

	record := c.clone()
	record.OID = catalog.GetController(id)
	record.created = time.Now()

	if auth != nil {
		if err := auth.CanAddController(record); err != nil {
			return nil, err
		}
	}

	cc.Controllers = append(cc.Controllers, record)

	return record, nil
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
		oid := catalog.GetController(k)

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
		Controllers: make([]*Controller, len(cc.Controllers)),
		LAN:         &LAN{},
	}

	for k, v := range cc.Controllers {
		shadow.Controllers[k] = v.clone()
	}

	shadow.LAN = cc.LAN.clone()

	return &shadow
}

func Export(file string, controllers []*Controller, doors map[string]doors.Door) error {
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
	cc.LAN.synchDoors(cc.Controllers)
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

func stringify(i interface{}) string {
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	default:
		return fmt.Sprintf("%v", i)
	}

	return ""
}
