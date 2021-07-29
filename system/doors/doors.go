package doors

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Doors struct {
	Doors map[string]Door `json:"doors"`
}

type object catalog.Object

var guard sync.Mutex

var trail audit.Trail

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func NewDoors() Doors {
	return Doors{
		Doors: map[string]Door{},
	}
}

func (dd *Doors) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, d := range dd.Doors {
		if d.IsValid() {
			if l := d.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (dd *Doors) Load(file string) error {
	created := time.Now()

	blob := struct {
		Doors []Door `json:"doors"`
	}{
		Doors: []Door{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	for _, d := range blob.Doors {
		d.created = created
		dd.Doors[d.OID] = d

		created = created.Add(1 * time.Second)
	}

	// for _, v := range cc.Controllers {
	//     if v.DeviceID != nil && *v.DeviceID != 0 {
	//         catalog.PutController(*v.DeviceID, v.OID)
	//     }
	// }

	return nil
}

// func (cc *ControllerSet) Save() error {
//     if cc == nil {
//         return nil
//     }

//     if err := validate(*cc); err != nil {
//         return err
//     }

//     if err := scrub(cc); err != nil {
//         return err
//     }

//     if cc.file == "" {
//         return nil
//     }

//     serializable := struct {
//         Controllers []json.RawMessage `json:"controllers"`
//         LAN         *LAN              `json:"LAN"`
//     }{
//         Controllers: []json.RawMessage{},
//         LAN:         cc.LAN.clone(),
//     }

//     for _, c := range cc.Controllers {
//         if record, err := c.serialize(); err == nil && record != nil {
//             fmt.Printf("                             >>>> %v\n", record)
//             serializable.Controllers = append(serializable.Controllers, record)
//         }
//     }

//     b, err := json.MarshalIndent(serializable, "", "  ")
//     if err != nil {
//         return err
//     }

//     tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted-controllers.json")
//     if err != nil {
//         return err
//     }

//     defer os.Remove(tmp.Name())

//     if _, err := tmp.Write(b); err != nil {
//         return err
//     }

//     if err := tmp.Close(); err != nil {
//         return err
//     }

//     if err := os.MkdirAll(filepath.Dir(cc.file), 0770); err != nil {
//         return err
//     }

//     return os.Rename(tmp.Name(), cc.file)
// }

// func (cc *ControllerSet) Sweep() {
//     if cc == nil {
//         return
//     }

//     cutoff := time.Now().Add(-cc.retention)
//     for i, v := range cc.Controllers {
//         if v.deleted != nil && v.deleted.Before(cutoff) {
//             cc.Controllers = append(cc.Controllers[:i], cc.Controllers[i+1:]...)
//         }
//     }
// }

func (dd *Doors) Print() {
	if dd != nil {
		if b, err := json.MarshalIndent(dd.Doors, "", "  "); err == nil {
			fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
		}
	}
}

func (dd *Doors) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if dd == nil {
		return nil, nil
	}

	for k, d := range dd.Doors {
		if strings.HasPrefix(oid, d.OID) {
			objects, err := d.set(auth, oid, value)
			if err == nil {
				dd.Doors[k] = d
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	//     if oid == "<new>" {
	//         if c, err := cc.add(auth, Controller{}); err != nil {
	//             return nil, err
	//         } else if c == nil {
	//             return nil, fmt.Errorf("Failed to add 'new' controller")
	//         } else {
	//             c.log(auth, "add", c.OID, "controller", "", "")
	//             objects = append(objects, object{
	//                 OID:   c.OID,
	//                 Value: "new",
	//             })
	//         }
	//     }

	return objects, nil
}

// func (cc *ControllerSet) add(auth auth.OpAuth, c Controller) (*Controller, error) {
//     id := uint32(0)
//     if c.DeviceID != nil {
//         id = *c.DeviceID
//     }

//     record := c.clone()
//     record.OID = catalog.Get(id)
//     record.created = time.Now()

//     if auth != nil {
//         if err := auth.CanAddController(record); err != nil {
//             return nil, err
//         }
//     }

//     cc.Controllers = append(cc.Controllers, record)

//     return record, nil
// }

// func (cc *ControllerSet) Refresh() {
//     cc.LAN.refresh(cc.Controllers)

//     // ... add 'found' controllers to list
// loop:
//     for k, _ := range cache.cache {
//         for _, c := range cc.Controllers {
//             if c.DeviceID != nil && *c.DeviceID == k && c.deleted == nil {
//                 continue loop
//             }
//         }

//         id := k
//         oid := catalog.Get(k)

//         cc.Controllers = append(cc.Controllers, &Controller{
//             OID:          oid,
//             DeviceID:     &id,
//             created:      time.Now(),
//             unconfigured: true,
//         })
//     }
// }

func (dd *Doors) Clone() *Doors {
	shadow := Doors{
		Doors: map[string]Door{},
	}

	for k, v := range dd.Doors {
		shadow.Doors[k] = v.clone()
	}

	return &shadow
}

// func Export(file string, controllers []*Controller, doors map[string]types.Door) error {
//     guard.Lock()

//     defer guard.Unlock()

//     conf := config.NewConfig()
//     if err := conf.Load(file); err != nil {
//         return err
//     }

//     devices := config.DeviceMap{}
//     for _, c := range controllers {
//         if c.DeviceID != nil && *c.DeviceID != 0 && c.deleted == nil {
//             device := config.Device{
//                 Name:     "",
//                 Address:  nil,
//                 Doors:    []string{"", "", "", ""},
//                 TimeZone: "",
//                 Rollover: 100000,
//             }

//             if c.Name != nil {
//                 device.Name = fmt.Sprintf("%v", c.Name)
//             }

//             if c.IP != nil {
//                 device.Address = (*net.UDPAddr)(c.IP)
//             }

//             if c.TimeZone != nil {
//                 device.TimeZone = *c.TimeZone
//             }

//             if d, ok := doors[c.Doors[1]]; ok {
//                 device.Doors[0] = d.Name
//             }

//             if d, ok := doors[c.Doors[2]]; ok {
//                 device.Doors[1] = d.Name
//             }

//             if d, ok := doors[c.Doors[3]]; ok {
//                 device.Doors[2] = d.Name
//             }

//             if d, ok := doors[c.Doors[4]]; ok {
//                 device.Doors[3] = d.Name
//             }

//             devices[*c.DeviceID] = &device
//         }
//     }

//     conf.Devices = devices

//     var b bytes.Buffer
//     conf.Write(&b)

//     tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted.conf_")
//     if err != nil {
//         return err
//     }

//     defer os.Remove(tmp.Name())

//     if _, err := tmp.Write(b.Bytes()); err != nil {
//         return err
//     }

//     if err := tmp.Close(); err != nil {
//         return err
//     }

//     if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
//         return err
//     }

//     return os.Rename(tmp.Name(), file)
// }

// func (cc *ControllerSet) Sync() {
//     cc.LAN.synchTime(cc.Controllers)
// }

// func (cc *ControllerSet) Compare(permissions acl.ACL) error {
//     return cc.LAN.compareACL(cc.Controllers, permissions)
// }

// func (cc *ControllerSet) UpdateACL(acl acl.ACL) {
//     cc.LAN.updateACL(cc.Controllers, acl)
// }

// func (cc *ControllerSet) Validate() error {
//     if cc != nil {
//         return validate(*cc)
//     }

//     return nil
// }

// func validate(cc ControllerSet) error {
//     devices := map[uint32]string{}

//     for _, c := range cc.Controllers {
//         if c.OID == "" {
//             return fmt.Errorf("Invalid controller OID (%v)", c.OID)
//         }

//         if c.deleted != nil {
//             continue
//         }

//         if c.DeviceID != nil && *c.DeviceID != 0 {
//             id := *c.DeviceID

//             if _, ok := devices[id]; ok {
//                 return fmt.Errorf("Duplicate controller ID (%v)", id)
//             }

//             devices[id] = c.OID
//         }
//     }

//     return nil
// }

// func scrub(cc *ControllerSet) error {
//     return nil
// }

// func warn(err error) {
//     log.Printf("ERROR %v", err)
// }

// func stringify(i interface{}) string {
//     switch v := i.(type) {
//     case *uint32:
//         if v != nil {
//             return fmt.Sprintf("%v", *v)
//         }

//     case *string:
//         if v != nil {
//             return fmt.Sprintf("%v", *v)
//         }

//     default:
//         return fmt.Sprintf("%v", i)
//     }

//     return ""
// }
