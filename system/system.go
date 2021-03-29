package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var sys = system{
	doors: struct {
		Doors map[string]types.Door `json:"doors"`
	}{
		Doors: map[string]types.Door{},
	},

	controllers: controllers.NewInterface(),
	taskQ:       NewTaskQ(),
}

type system struct {
	sync.RWMutex
	conf  string
	doors struct {
		Doors map[string]types.Door `json:"doors"`
	}
	controllers controllers.Interface
	cards       cards.Cards
	audit       audit.Trail
	taskQ       TaskQ
}

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

func Init(conf, controllers, doors string, cards cards.Cards, trail audit.Trail, retention time.Duration) error {
	sys.controllers.Load(controllers, retention)

	bytes, err := ioutil.ReadFile(doors)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &sys.doors)
	if err != nil {
		return err
	}

	sys.conf = conf
	sys.cards = cards
	sys.audit = trail

	sys.controllers.Print()
	//	if b, err := json.MarshalIndent(sys.doors, "", "  "); err == nil {
	//		fmt.Printf("-----------------\n%s\n-----------------\n", string(b))
	//	}

	return nil
}

func System() interface{} {
	sys.RLock()

	defer sys.RUnlock()

	controllers := controllers.Consolidate(sys.controllers.LAN, sys.controllers.Interface)

	doors := []types.Door{}
	for _, v := range sys.doors.Doors {
		doors = append(doors, v)
	}

	sort.SliceStable(doors, func(i, j int) bool { return doors[i].Name < doors[j].Name })

	return struct {
		Controllers interface{}
		Doors       []types.Door
	}{
		Controllers: controllers,
		Doors:       doors,
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
		sys.controllers.LAN.Update(*acl)
	}
}

func CompareACL() {
	if permissions, err := sys.cards.ACL(); err != nil {
		warn(err)
	} else if acl, err := consolidate(permissions); err != nil {
		warn(err)
	} else if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
	} else if err := sys.controllers.LAN.Compare(*acl); err != nil {
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

	for _, c := range sys.controllers.Interface {
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
			door, ok := sys.doors.Doors[d]
			if !ok {
				log.Printf("WARN %v", fmt.Errorf("consolidate: invalid door %v for card %v", d, p.CardNumber))
				continue
			}

			for _, c := range sys.controllers.Interface {
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

func info(msg string) {
	log.Printf("INFO  %v", msg)
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
