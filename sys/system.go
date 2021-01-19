package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type system struct {
	sync.RWMutex
	file string
	data data
}

type data struct {
	Tables tables `json:"tables"`
}

type tables struct {
	Doors       map[string]types.Door  `json:"doors"`
	Controllers map[string]*Controller `json:"controllers"`
	Local       *Local                 `json:"local"`
}

func (s *system) refresh() {
	if s != nil {
		go s.data.Tables.Local.refresh()
	}
}

func (d *data) clone() *data {
	shadow := data{
		Tables: tables{
			Doors:       map[string]types.Door{},
			Controllers: map[string]*Controller{},
			Local:       &Local{},
		},
	}

	for k, v := range d.Tables.Doors {
		shadow.Tables.Doors[k] = v.Clone()
	}

	for k, v := range d.Tables.Controllers {
		shadow.Tables.Controllers[k] = v.clone()
	}

	shadow.Tables.Local = d.Tables.Local.clone()

	return &shadow
}

var sys = system{
	data: data{
		Tables: tables{
			Doors:       map[string]types.Door{},
			Controllers: map[string]*Controller{},
			Local: &Local{
				devices: map[uint32]address{},
			},
		},
	},
}

func init() {
	go func() {
		time.Sleep(2500 * time.Millisecond)
		sys.refresh()

		c := time.Tick(30 * time.Second)
		for _ = range c {
			sys.refresh()
		}
	}()
}

func Init(conf string) error {
	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &sys.data)
	if err != nil {
		return err
	}

	sys.file = conf

	sys.data.Tables.Local.Init(sys.data.Tables.Controllers)

	//	if b, err := json.MarshalIndent(sys.data, "", "  "); err == nil {
	//		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	//	}

	return nil
}

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	devices := []Controller{}
	for _, v := range sys.data.Tables.Controllers {
		devices = append(devices, *v)
	}

loop:
	for k, _ := range sys.data.Tables.Local.cache {
		for _, c := range devices {
			if c.DeviceID != nil && *c.DeviceID == k {
				continue loop
			}
		}

		id := k
		devices = append(devices, Controller{
			ID:       ID(k),
			DeviceID: &id,
			Created:  time.Now(),
		})
	}

	controllers := []controller{}
	for _, c := range devices {
		controllers = append(controllers, merge(c))
	}

	sort.SliceStable(controllers, func(i, j int) bool { return controllers[i].Created.Before(controllers[j].Created) })

	return struct {
		Controllers []controller
	}{
		Controllers: controllers,
	}
}

func Update(permissions []types.Permissions) {
	sys.data.Tables.Local.Update(permissions)
}

//func Post(m map[string]interface{}, auth db.IAuth) (interface{}, error) {
func Post(m map[string]interface{}) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	// add/update ?

	controllers, err := unpack(m)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid request (%v)", err),
			Detail: fmt.Errorf("Error unpacking 'post' request (%w)", err),
		}
	}

	list := struct {
		Updated []interface{} `json:"updated"`
		Deleted []interface{} `json:"deleted"`
	}{}

	shadow := sys.data.clone()

loop:
	for _, c := range controllers {
		if c.OID == "" {
			if r, err := add(shadow, c); err != nil {
				return nil, err
			} else if r != nil {
				list.Updated = append(list.Updated, merge(*r))
			}

			continue loop
		}

		if c.ID == "" {
			return nil, &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid controller ID"),
				Detail: fmt.Errorf("Invalid 'post' request (%w)", fmt.Errorf("Invalid controller ID '%v'", c.ID)),
			}
		}

		// if c.Name != nil && *c.Name == "" && c.Card != nil && *c.Card == 0 {
		// 	if r, err := d.delete(shadow, c, auth); err != nil {
		// 		return nil, err
		// 	} else if r != nil {
		// 		list.Deleted = append(list.Deleted, r)
		// 		continue loop
		// 	}
		// }

		if r, err := update(shadow, c); err != nil {
			return nil, err
		} else if r != nil {
			list.Updated = append(list.Updated, merge(*r))
		}

		continue loop
	}

	if err := save(shadow, sys.file); err != nil {
		return nil, err
	}

	sys.data = *shadow

	return list, nil
}

//func add(shadow *data, ch types.CardHolder, auth db.IAuth) (interface{}, error) {
func add(shadow *data, c Controller) (*Controller, error) {
	record := c.clone()

	//	if auth != nil {
	//		if err := auth.CanAddCardHolder(record); err != nil {
	//			return nil, &types.HttpdError{
	//				Status: http.StatusUnauthorized,
	//				Err:    fmt.Errorf("Not authorized to add card holder"),
	//				Detail: err,
	//			}
	//		}
	//	}

loop:
	for next := 1; ; next++ {
		oid := fmt.Sprintf("0.1.1.%v", next)
		for _, v := range shadow.Tables.Controllers {
			if v.OID == oid {
				continue loop
			}
		}

		record.OID = oid
		break
	}

	record.Created = time.Now()

	shadow.Tables.Controllers[record.ID] = record
	//	d.log("add", record, auth)

	return record, nil
}

// func update(shadow *data, ch types.CardHolder, auth db.IAuth) (interface{}, error) {
func update(shadow *data, c Controller) (*Controller, error) {
	for _, record := range shadow.Tables.Controllers {
		if record.OID == c.OID {
			if c.Name != nil {
				record.Name = c.Name
			}

			if c.DeviceID != nil && *c.DeviceID != 0 {
				id := *c.DeviceID
				record.DeviceID = &id
			}

			if c.IP != nil {
				record.IP = c.IP.clone()
			}

			//		current := d.data.Tables.CardHolders[ch.ID]
			//		if auth != nil {
			//			if err := auth.CanUpdateCardHolder(current, record); err != nil {
			//				return nil, &types.HttpdError{
			//					Status: http.StatusUnauthorized,
			//					Err:    fmt.Errorf("Not authorized to update card holder"),
			//					Detail: err,
			//				}
			//			}
			//		}
			//
			//		d.log("update", map[string]interface{}{"original": current, "updated": record}, auth)

			return record, nil
		}
	}

	return nil, &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid controller OID"),
		Detail: fmt.Errorf("Invalid 'post' request (%w)", fmt.Errorf("Invalid controller OID '%v'", c.OID)),
	}
}

func save(d *data, file string) error {
	if err := validate(d); err != nil {
		return err
	}

	if err := scrub(d); err != nil {
		return err
	}

	if file == "" {
		return nil
	}

	_, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	//	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted-system.json")
	//	if err != nil {
	//		return err
	//	}
	//
	//	defer os.Remove(tmp.Name())
	//
	//	if _, err := tmp.Write(b); err != nil {
	//		return err
	//	}
	//
	//	if err := tmp.Close(); err != nil {
	//		return err
	//	}
	//
	//	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
	//		return err
	//	}
	//
	//	return os.Rename(tmp.Name(), file)

	return nil
}

func validate(d *data) error {
	devices := map[uint32]string{}

	for _, r := range d.Tables.Controllers {

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
					Detail: fmt.Errorf("controller %v: duplicate device ID in records %v and %v", id, rid, r.ID),
				}
			}

			devices[id] = r.ID
		}
	}

	return nil
}

func scrub(d *data) error {
	return nil
}

func consolidate(list []types.Permissions) (*acl.ACL, error) {
	// initialise empty ACL
	acl := make(acl.ACL)

	for _, d := range sys.data.Tables.Doors {
		if _, ok := acl[d.DeviceID]; !ok {
			acl[d.DeviceID] = make(map[uint32]core.Card)
		}
	}

	// create ACL with all cards on all controllers
	for _, p := range list {
		for _, l := range acl {
			if _, ok := l[p.CardNumber]; !ok {
				from := core.Date(p.From)
				to := core.Date(p.To)

				l[p.CardNumber] = core.Card{
					CardNumber: p.CardNumber,
					From:       &from,
					To:         &to,
					Doors:      map[uint8]bool{1: false, 2: false, 3: false, 4: false},
				}
			}
		}
	}

	// update ACL cards from permissions
	for _, p := range list {
		for _, d := range p.Doors {
			if door, ok := sys.data.Tables.Doors[d]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Invalid door %v for card %v", d, p.CardNumber))
			} else if l, ok := acl[door.DeviceID]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Door %v - invalid configuration (no controller defined for  %v)", d, door.DeviceID))
			} else if card, ok := l[p.CardNumber]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Card %v not initialised for controller %v", p.CardNumber, door.DeviceID))
			} else {
				card.Doors[door.Door] = true
			}
		}
	}

	return &acl, nil
}

func unpack(m map[string]interface{}) ([]Controller, error) {
	o := struct {
		Controllers []struct {
			ID       string
			OID      *string
			Name     *string
			DeviceID *uint32
			IP       *string
		}
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	fmt.Printf(">> DEBUG: %s\n", string(blob))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, err
	}

	controllers := []Controller{}

	for _, r := range o.Controllers {
		record := Controller{
			ID: strings.TrimSpace(r.ID),
		}

		if r.OID != nil {
			record.OID = *r.OID
		}

		if r.Name != nil {
			name := types.Name(*r.Name)
			record.Name = &name
		}

		if r.DeviceID != nil {
			record.DeviceID = r.DeviceID
		}

		if r.IP != nil {
			if addr, err := resolve(*r.IP); err != nil {
				return nil, err
			} else {
				record.IP = addr
			}
		}

		controllers = append(controllers, record)
	}

	return controllers, nil
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
