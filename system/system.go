package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

var sys = system{
	doors:       doors.NewDoors(),
	controllers: controllers.NewControllerSet(),
	taskQ:       NewTaskQ(),
}

var resolver = Resolver{}

type system struct {
	sync.RWMutex
	conf        string
	doors       doors.Doors
	controllers controllers.ControllerSet
	cards       cards.Cards
	audit       audit.Trail
	taskQ       TaskQ
}

type object catalog.Object

func (s *system) refresh() {
	if s == nil {
		return
	}

	sys.taskQ.Add(Task{
		f: s.controllers.Refresh,
	})

	sys.taskQ.Add(Task{
		f: s.controllers.Sweep,
	})

	sys.taskQ.Add(Task{
		f: CompareACL,
	})
}

func Init(cfg config.Config, conf string, cards cards.Cards, trail audit.Trail, retention time.Duration) error {
	catalog.SetResolver(resolver)

	if err := sys.doors.Load(cfg.HTTPD.System.Doors); err != nil {
		return err
	}

	if err := sys.controllers.Load(cfg.HTTPD.System.Controllers, retention); err != nil {
		return err
	}

	sys.conf = conf
	sys.cards = cards
	sys.audit = trail

	controllers.SetAuditTrail(trail)
	doors.SetAuditTrail(trail)

	sys.controllers.Print()
	sys.doors.Print()

	go func() {
		time.Sleep(2500 * time.Millisecond)
		sys.refresh()

		c := time.Tick(cfg.HTTPD.System.Refresh)
		for _ = range c {
			sys.refresh()
		}
	}()

	return nil
}

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	objects := []interface{}{}
	objects = append(objects, sys.controllers.AsObjects()...)
	objects = append(objects, sys.doors.AsObjects()...)

	d := []doors.Door{}
	for _, v := range sys.doors.Doors {
		if v.IsValid() {
			d = append(d, v)
		}
	}

	sort.SliceStable(d, func(i, j int) bool { return d[i].Name < d[j].Name })

	return struct {
		Objects []interface{} `json:"objects"`
		Doors   []doors.Door  `json:"doors"`
	}{
		Objects: objects,
		Doors:   d,
	}
}

func Cards() interface{} {
	return sys.cards.CardHolders()
}

func Groups() interface{} {
	return sys.cards.Groups()
}

func UpdateACL() {
	if permissions, err := sys.cards.ACL(); err != nil {
		warn(err)
	} else if acl, err := consolidate(permissions); err != nil {
		warn(err)
	} else if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
	} else {
		sys.controllers.UpdateACL(*acl)
	}
}

func CompareACL() {
	if permissions, err := sys.cards.ACL(); err != nil {
		warn(err)
	} else if acl, err := consolidate(permissions); err != nil {
		warn(err)
	} else if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
	} else if err := sys.controllers.Compare(*acl); err != nil {
		warn(err)
	}
}

func UpdateCardHolders(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	response, err := sys.cards.Post(m, auth)
	if err != nil {
		return nil, err
	}

	sys.taskQ.Add(Task{
		f: UpdateACL,
	})

	return response, nil
}

func consolidate(list []types.Permissions) (*acl.ACL, error) {
	// initialise empty ACL
	acl := make(acl.ACL)

	for _, c := range sys.controllers.Controllers {
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
					Doors:      map[uint8]int{1: 0, 2: 0, 3: 0, 4: 0},
				}
			}
		}
	}

	// update ACL cards from permissions

	for _, p := range list {
	loop:
		for _, d := range p.Doors {
			door, ok := sys.doors.Doors[d]
			if !ok {
				log.Printf("WARN %v", fmt.Errorf("consolidate: invalid door %v for card %v", d, p.CardNumber))
				continue
			}

			for _, c := range sys.controllers.Controllers {
				for _, v := range c.Doors {
					if v == door.OID {
						if c.DeviceID != nil && *c.DeviceID > 0 {
							// FIXME: reinstate once the 'doors' rework is done
							// if l, ok := acl[*c.DeviceID]; ok {
							// 	if card, ok := l[p.CardNumber]; !ok {
							// 		log.Printf("WARN %v", fmt.Errorf("consolidate: card %v not initialised for controller %v", p.CardNumber, *c.DeviceID))
							// 	} else {
							// 		card.Doors[door.Door] = 1
							// 	}
							// }
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

func unpack(m map[string]interface{}) ([]object, error) {
	f := func(err error) error {
		return types.BadRequest(fmt.Errorf("Invalid request (%v)", err), fmt.Errorf("Error unpacking 'post' request (%w)", err))
	}

	o := struct {
		Objects []object `json:"objects"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, f(err)
	}

	log.Printf("DEBUG %v", fmt.Sprintf("UNPACK %s\n", string(blob)))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, f(err)
	}

	return o.Objects, nil
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func info(msg string) {
	log.Printf("INFO  %v", msg)
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
