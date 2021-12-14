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
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

const BLANK = "'blank'"

var guard sync.RWMutex

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

func SetWindows(ok, uncertain, systime, cacheExpiry time.Duration) {
	windows.deviceOk = ok
	windows.deviceUncertain = uncertain
	windows.systime = systime
	windows.cacheExpiry = cacheExpiry
}

func NewControllers() Controllers {
	return Controllers{
		Controllers: []*Controller{},
	}
}

func (cc *Controllers) Init(interfaces interfaces.Interfaces) {
	for _, v := range interfaces.LANs {
		if v != nil {
			cc.LAN = &LAN{
				*v,
			}

			break
		}
	}
}

func (cc *Controllers) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	cc.Controllers = []*Controller{}
	for _, v := range rs {
		var c Controller
		if err := c.deserialize(v); err != nil {
			warn(err)
		} else {
			cc.Controllers = append(cc.Controllers, &c)
		}
	}

	for _, c := range cc.Controllers {
		if c.deviceID != nil && *c.deviceID != 0 {
			oid := c.OID()
			catalog.PutController(*c.deviceID, oid)
			catalog.PutV(oid, ControllerName, c.name)
			catalog.PutV(oid, ControllerDoor1, c.Doors[1])
			catalog.PutV(oid, ControllerDoor2, c.Doors[2])
			catalog.PutV(oid, ControllerDoor3, c.Doors[3])
			catalog.PutV(oid, ControllerDoor4, c.Doors[4])
		}
	}

	return nil
}

func (cc *Controllers) Save() (json.RawMessage, error) {
	if cc == nil {
		return nil, nil
	}

	if err := validate(*cc); err != nil {
		return nil, err
	}

	if err := scrub(cc); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, c := range cc.Controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (cc *Controllers) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, c := range cc.Controllers {
		if c.IsValid() {
			if l := c.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (cc *Controllers) Sweep(retention time.Duration) {
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

func (cc *Controllers) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("----------------- CONTROLLERS\n%s\n", string(b))
	}
}

func (cc *Controllers) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if cc == nil {
		return nil, nil
	}

	for _, c := range cc.Controllers {
		if c != nil && c.OID().Contains(oid) {
			return c.set(auth, oid, value, dbc)
		}
	}

	objects := []catalog.Object{}

	if oid == "<new>" {
		if c, err := cc.add(auth, Controller{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			OID := c.OID()
			c.log(auth, "add", OID, "controller", fmt.Sprintf("Added 'new' controller"), "", "", dbc)
			objects = append(objects, catalog.NewObject(OID, "new"))
			objects = append(objects, catalog.NewObject2(OID, ControllerStatus, "new"))
			objects = append(objects, catalog.NewObject2(OID, ControllerCreated, c.created))
		}
	}

	return objects, nil
}

func (cc *Controllers) Refresh() {
	// ... add 'found' controllers to list
	if found, err := cc.LAN.search(cc.Controllers); err != nil {
		warn(err)
	} else {
	loop:
		for _, d := range found {
			for _, c := range cc.Controllers {
				if c.DeviceID() == d && c.deleted == nil {
					continue loop
				}
			}

			info(fmt.Sprintf("Adding unconfigured controller %v", d))

			oid := catalog.NewController(d)
			deviceID := d // because .. Go loop variable gotcha (the loop variable is mutable)

			cc.Controllers = append(cc.Controllers, &Controller{
				oid:          oid,
				deviceID:     &deviceID,
				created:      types.DateTime(time.Now()),
				unconfigured: true,
			})
		}
	}

	// ... refresh
	cc.LAN.refresh(cc.Controllers)

	for _, c := range cc.Controllers {
		c.refreshed()
	}
}

func (cc *Controllers) Clone() Controllers {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Controllers{
		Controllers: make([]*Controller, len(cc.Controllers)),
		LAN:         &LAN{},
	}

	lan := cc.LAN.clone()
	shadow.LAN = &lan

	for k, v := range cc.Controllers {
		shadow.Controllers[k] = v.clone()
	}

	return shadow
}

func Export(file string, controllers []*Controller, doors map[catalog.OID]doors.Door) error {
	guard.RLock()

	defer guard.RUnlock()

	conf := config.NewConfig()
	if err := conf.Load(file); err != nil {
		return err
	}

	devices := config.DeviceMap{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			device := config.Device{
				Name:     c.name,
				Address:  nil,
				Doors:    []string{"", "", "", ""},
				TimeZone: "",
				Rollover: 100000,
			}

			if c.IP != nil {
				device.Address = (*net.UDPAddr)(c.IP)
			}

			if c.timezone != nil {
				device.TimeZone = *c.timezone
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

			devices[*c.deviceID] = &device
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
	cc.LAN.synchTime(cc.Controllers)
	cc.LAN.synchDoors(cc.Controllers)
}

func (cc *Controllers) CompareACL(permissions acl.ACL) error {
	return cc.LAN.compare(cc.Controllers, permissions)
}

func (cc *Controllers) UpdateACL(permissions acl.ACL) error {
	return cc.LAN.update(cc.Controllers, permissions)
}

func (cc *Controllers) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func (cc *Controllers) add(auth auth.OpAuth, c Controller) (*Controller, error) {
	id := uint32(0)
	if c.deviceID != nil {
		id = *c.deviceID
	}

	record := c.clone()
	record.oid = catalog.OID(catalog.NewController(id))
	record.created = types.DateTime(time.Now())

	if auth != nil {
		if err := auth.CanAddController(record); err != nil {
			return nil, err
		}
	}

	cc.Controllers = append(cc.Controllers, record)

	return record, nil
}

func validate(cc Controllers) error {
	devices := map[uint32]string{}

	for _, c := range cc.Controllers {
		OID := c.OID()
		if OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", OID)
		}

		if c.deleted != nil {
			continue
		}

		if c.deviceID != nil && *c.deviceID != 0 {
			id := *c.deviceID

			if _, ok := devices[id]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", id)
			}

			devices[id] = string(OID)
		}
	}

	return nil
}

func scrub(cc *Controllers) error {
	return nil
}

func info(msg string) {
	log.Printf("INFO  %v", msg)
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}

func stringify(i interface{}, defval string) string {
	s := ""
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	default:
		s = fmt.Sprintf("%v", i)
	}

	if s != "" {
		return s
	}

	return defval
}
