package system

import (
	"bytes"
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
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type system struct {
	sync.RWMutex
	file  string
	data  data
	audit audit.Trail
}

type data struct {
	Tables tables `json:"tables"`
}

type tables struct {
	Doors       map[string]types.Door `json:"doors"`
	Controllers []*Controller         `json:"controllers"`
	Local       *Local                `json:"local"`
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
			Controllers: make([]*Controller, len(d.Tables.Controllers)),
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
			Controllers: []*Controller{},
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

func Init(conf string, trail audit.Trail) error {
	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &sys.data)
	if err != nil {
		return err
	}

	sys.audit = trail
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
			DeviceID: &id,
			Created:  time.Now(),
		})
	}

	controllers := []controller{}
	for _, c := range devices {
		controllers = append(controllers, merge(c))
	}

	sort.SliceStable(controllers, func(i, j int) bool { return controllers[i].Created.Before(controllers[j].Created) })

	doors := []types.Door{}
	for _, v := range sys.data.Tables.Doors {
		doors = append(doors, v)
	}

	sort.SliceStable(doors, func(i, j int) bool { return doors[i].Name < doors[j].Name })

	return struct {
		Controllers []controller
		Doors       []types.Door
	}{
		Controllers: controllers,
		Doors:       doors,
	}
}

func Update(permissions []types.Permissions) {
	sys.data.Tables.Local.Update(permissions)
}

func Post(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
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
		// ... delete?
		if (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
			for _, v := range shadow.Tables.Controllers {
				if v.OID == c.OID {
					if r, err := sys.delete(shadow, c, auth); err != nil {
						return nil, err
					} else if r != nil {
						list.Deleted = append(list.Deleted, merge(*r))
					}
				}
			}

			continue loop
		}

		// ... update controller?
		for _, v := range shadow.Tables.Controllers {
			if v.OID == c.OID {
				if r, err := sys.update(shadow, c, auth); err != nil {
					return nil, err
				} else if r != nil {
					list.Updated = append(list.Updated, merge(*r))
				}

				continue loop
			}
		}

		// ... add controller
		if r, err := sys.add(shadow, c, auth); err != nil {
			return nil, err
		} else if r != nil {
			list.Updated = append(list.Updated, merge(*r))
		}
	}

	if err := save(shadow, sys.file); err != nil {
		return nil, err
	}

	sys.data = *shadow

	return list, nil
}

func (s *system) add(shadow *data, c Controller, auth auth.OpAuth) (*Controller, error) {
	record := c.clone()

	if auth != nil {
		if err := auth.CanAddController(record); err != nil {
			return nil, &types.HttpdError{
				Status: http.StatusUnauthorized,
				Err:    fmt.Errorf("Not authorized to add controller"),
				Detail: err,
			}
		}
	}

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

	shadow.Tables.Controllers = append(shadow.Tables.Controllers, record)
	s.log("add", record, auth)

	return record, nil
}

func (s *system) update(shadow *data, c Controller, auth auth.OpAuth) (*Controller, error) {
	var current *Controller

	for _, v := range s.data.Tables.Controllers {
		if v.OID == c.OID {
			current = v
			break
		}
	}

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

			if c.TimeZone != nil {
				tz := *c.TimeZone
				record.TimeZone = &tz
			}

			if c.Doors != nil {
				for k, v := range c.Doors {
					record.Doors[k] = v
				}
			}

			if auth != nil {
				if err := auth.CanUpdateController(current, record); err != nil {
					return nil, &types.HttpdError{
						Status: http.StatusUnauthorized,
						Err:    fmt.Errorf("Not authorized to update controller"),
						Detail: err,
					}
				}
			}

			s.log("update", map[string]interface{}{"original": current, "updated": record}, auth)

			return record, nil
		}
	}

	return nil, &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid controller OID"),
		Detail: fmt.Errorf("Invalid 'post' request (%w)", fmt.Errorf("Invalid controller OID '%v'", c.OID)),
	}
}

func (s *system) delete(shadow *data, c Controller, auth auth.OpAuth) (*Controller, error) {
	for i, record := range shadow.Tables.Controllers {
		if record.OID == c.OID {
			if auth != nil {
				if err := auth.CanDeleteController(record); err != nil {
					return nil, &types.HttpdError{
						Status: http.StatusUnauthorized,
						Err:    fmt.Errorf("Not authorized to delete controller"),
						Detail: err,
					}
				}
			}

			shadow.Tables.Controllers = append(shadow.Tables.Controllers[:i], shadow.Tables.Controllers[i+1:]...)

			s.log("delete", record, auth)

			return &c, nil
		}
	}

	return nil, nil
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
	doors := map[string]string{}

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
					Detail: fmt.Errorf("controller %v: duplicate device ID in records %v and %v", id, rid, r.OID),
				}
			}

			devices[id] = r.OID
		}

		for _, v := range r.Doors {
			if v != "" {
				if _, ok := d.Tables.Doors[v]; !ok {
					return &types.HttpdError{
						Status: http.StatusBadRequest,
						Err:    fmt.Errorf("Invalid door ID"),
						Detail: fmt.Errorf("controller %v: invalid door ID (%v)", r.OID, v),
					}
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("%v door assigned to more than one controller", d.Tables.Doors[v].Name),
					Detail: fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID),
				}
			}

			doors[v] = r.OID
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

	for _, c := range sys.data.Tables.Controllers {
		if c.DeviceID != nil && *c.DeviceID > 0 {
			acl[*c.DeviceID] = map[uint32]core.Card{}
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
	loop:
		for _, d := range p.Doors {
			door, ok := sys.data.Tables.Doors[d]
			if !ok {
				log.Printf("WARN %v", fmt.Errorf("consolidate: invalid door %v for card %v", d, p.CardNumber))
				continue
			}

			for _, c := range sys.data.Tables.Controllers {
				for _, v := range c.Doors {
					if v == door.ID {
						if c.DeviceID != nil && *c.DeviceID > 0 {
							if l, ok := acl[*c.DeviceID]; ok {
								if card, ok := l[p.CardNumber]; !ok {
									log.Printf("WARN %v", fmt.Errorf("consolidate: card %v not initialised for controller %v", p.CardNumber, *c.DeviceID))
								} else {
									card.Doors[door.Door] = true
								}
							}
						}

						continue loop
					}
				}
			}

			log.Printf("WARN %v", fmt.Errorf("consolidate: card %v, door %v - no controller assigned", p.CardNumber, door))
		}
	}

	var b bytes.Buffer

	acl.Print(&b)
	log.Printf("INFO %v", fmt.Sprintf("ACL\n%s", string(b.Bytes())))

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
			Doors    map[uint8]string
			DateTime *string
		}
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, err
	}

	controllers := []Controller{}

	for _, r := range o.Controllers {
		record := Controller{}

		record.ID = r.ID

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

		if r.DateTime != nil {
			if tz, err := timezone(strings.TrimSpace(*r.DateTime)); err != nil {
				return nil, err
			} else {
				tzs := tz.String()
				record.TimeZone = &tzs
			}
		}

		if r.Doors != nil && len(r.Doors) > 0 {
			record.Doors = map[uint8]string{}
			for k, v := range r.Doors {
				record.Doors[k] = v
			}
		}

		controllers = append(controllers, record)
	}

	return controllers, nil
}

func (s *system) log(op string, info interface{}, auth auth.OpAuth) {
	if s.audit != nil {
		uid := ""
		if auth != nil {
			uid = auth.UID()
		}

		s.audit.Write(audit.LogEntry{
			UID:       uid,
			Module:    "system",
			Operation: op,
			Info:      info,
		})
	}
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
