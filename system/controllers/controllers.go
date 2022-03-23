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
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	controllers []*Controller
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
		controllers: []*Controller{},
	}
}

func (cc *Controllers) AsObjects(auth auth.OpAuth) []schema.Object {
	objects := []schema.Object{}

	for _, c := range cc.controllers {
		if c.IsValid() {
			catalog.Join(&objects, c.AsObjects(auth)...)
		}
	}

	return objects
}

func (cc *Controllers) UpdateByOID(auth auth.OpAuth, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if cc == nil {
		return nil, nil
	}

	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	for _, c := range cc.controllers {
		if c != nil && c.OID().Contains(oid) {
			return c.set(auth, oid, value, dbc)
		}
	}

	objects := []schema.Object{}

	if oid == "<new>" {
		if c, err := cc.add(auth, Controller{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			OID := c.OID()

			catalog.Join(&objects, catalog.NewObject(OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(OID, ControllerStatus, "new"))
			catalog.Join(&objects, catalog.NewObject2(OID, ControllerCreated, c.created))

			c.log(uid, "add", OID, "controller", fmt.Sprintf("Added 'new' controller"), "", "", dbc)
		}
	}

	return objects, nil
}

func (cc *Controllers) List() []Controller {
	list := []Controller{}

	for _, c := range cc.controllers {
		if c != nil {
			list = append(list, *c)
		}
	}

	return list
}

func (cc *Controllers) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	cc.controllers = []*Controller{}
	for _, v := range rs {
		var c Controller
		if err := c.deserialize(v); err != nil {
			warn(err)
		} else {
			cc.controllers = append(cc.controllers, &c)
		}
	}

	for _, c := range cc.controllers {
		oid := c.OID()
		catalog.PutT(c.CatalogController, oid)
		catalog.PutV(oid, ControllerName, c.name)
		catalog.PutV(oid, ControllerDeviceID, c.DeviceID)
		catalog.PutV(oid, ControllerDateTimeModified, false)
		catalog.PutV(oid, ControllerDoor1, c.Doors[1])
		catalog.PutV(oid, ControllerDoor2, c.Doors[2])
		catalog.PutV(oid, ControllerDoor3, c.Doors[3])
		catalog.PutV(oid, ControllerDoor4, c.Doors[4])
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
	for _, c := range cc.controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (cc *Controllers) Sweep(retention time.Duration) {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-retention)
	for i, v := range cc.controllers {
		if v.IsDeleted() && v.deleted.Before(cutoff) {
			cc.controllers = append(cc.controllers[:i], cc.controllers[i+1:]...)
		}
	}
}

func (cc Controllers) Print() {
	serializable := []json.RawMessage{}
	for _, c := range cc.controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- CONTROLLERS\n%s\n", string(b))
	}
}

func (cc *Controllers) Refresh(i interfaces.Interfaces) {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return
	} else {
		lan = LAN{
			v,
		}
	}

	// ... add 'found' controllers to list
	if found, err := lan.search(cc.controllers); err != nil {
		warn(err)
	} else {
	loop:
		for _, d := range found {
			id := d // because .. Go loop variable gotcha (the loop variable is mutable)
			for _, c := range cc.controllers {
				if c.ID() == id && !c.IsDeleted() {
					continue loop
				}
			}

			info(fmt.Sprintf("Adding unconfigured controller %v", d))

			v := Controller{
				CatalogController: ctypes.CatalogController{
					DeviceID: id,
				},
				created:      types.TimestampNow(),
				unconfigured: true,
			}

			v.oid = catalog.NewT(v.CatalogController)

			cc.controllers = append(cc.controllers, &v)
		}
	}

	// ... refresh
	lan.refresh(cc.controllers)

	for _, c := range cc.controllers {
		c.refreshed()
	}
}

func (cc *Controllers) Clone() Controllers {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Controllers{
		controllers: make([]*Controller, len(cc.controllers)),
	}

	for k, v := range cc.controllers {
		shadow.controllers[k] = v.clone()
	}

	return shadow
}

func Export(file string, controllers []*Controller, doors map[schema.OID]doors.Door) error {
	guard.RLock()

	defer guard.RUnlock()

	conf := config.NewConfig()
	if err := conf.Load(file); err != nil {
		return err
	}

	devices := config.DeviceMap{}
	for _, c := range controllers {
		if c.realized() {
			device := config.Device{
				Name:     c.name,
				Address:  nil,
				Doors:    []string{"", "", "", ""},
				TimeZone: c.timezone,
			}

			if c.IP != nil {
				device.Address = (*net.UDPAddr)(c.IP)
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

			devices[c.DeviceID] = &device
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

func (cc *Controllers) Sync(i interfaces.Interfaces) {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return
	} else {
		lan = LAN{
			v,
		}
	}

	lan.synchTime(cc.controllers)
	lan.synchDoors(cc.controllers)
}

func (cc *Controllers) CompareACL(i interfaces.Interfaces, permissions acl.ACL) error {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return fmt.Errorf("No active LAN subsystem")
	} else {
		lan = LAN{
			v,
		}
	}

	return lan.compare(cc.controllers, permissions)
}

func (cc *Controllers) UpdateACL(i interfaces.Interfaces, permissions acl.ACL) error {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return fmt.Errorf("No active LAN subsystem")
	} else {
		lan = LAN{
			v,
		}
	}

	return lan.update(cc.controllers, permissions)
}

func (cc *Controllers) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func (cc *Controllers) add(a auth.OpAuth, c Controller) (*Controller, error) {
	record := c.clone()
	record.oid = schema.OID(catalog.NewT(c.CatalogController))
	record.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(record, auth.Controllers); err != nil {
			return nil, err
		}
	}

	cc.controllers = append(cc.controllers, record)

	return record, nil
}

func validate(cc Controllers) error {
	devices := map[uint32]string{}

	for _, c := range cc.controllers {
		OID := c.OID()
		if OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", OID)
		}

		if c.IsDeleted() {
			continue
		}

		if c.DeviceID != 0 {
			if _, ok := devices[c.DeviceID]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", c.DeviceID)
			}

			devices[c.DeviceID] = string(OID)
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
