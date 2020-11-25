package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type system struct {
	Controllers []Controller          `json:"controllers"`
	Doors       map[string]types.Door `json:"doors"`
	Local       []Local               `json:"local"`
}

type Controller struct {
	Name     string         `json:"name"`
	ID       uint32         `json:"id"`
	IP       string         `json:"ip"`
	DateTime types.DateTime `json:"datetime"`
	Cards    uint32         `json:"cards"`
	Events   uint32         `json:"events"`
	Doors    map[int]string `json:"doors"`
}

var sys = system{
	Controllers: []Controller{},
	Doors:       map[string]types.Door{},
	Local:       []Local{},
}

func resolve(address string) *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", address)

	return addr
}

func Init(conf string) error {
	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &sys)
	if err != nil {
		return err
	}

	return nil
}

func System() interface{} {
	return struct {
		Controllers []Controller
	}{
		Controllers: []Controller{
			Controller{
				Name:     "Top",
				ID:       12345678,
				IP:       "192.168.1.100:60000",
				DateTime: types.DateTime(time.Now()),
				Cards:    17,
				Events:   29,
				Doors: map[int]string{
					1: "D1",
					2: "D2",
					3: "D3",
					4: "D4",
				},
			},
		},
	}
}

func Update(permissions []types.Permissions) {
	for _, l := range sys.Local {
		l.Update(permissions)
	}
}

func consolidate(list []types.Permissions) (*acl.ACL, error) {
	// initialise empty ACL
	acl := make(acl.ACL)

	for _, d := range sys.Doors {
		if _, ok := acl[d.ControllerID]; !ok {
			acl[d.ControllerID] = make(map[uint32]core.Card)
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
			if door, ok := sys.Doors[d]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Invalid door %v for card %v", d, p.CardNumber))
			} else if l, ok := acl[door.ControllerID]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Door %v - invalid configuration (no controller defined for  %v)", d, door.ControllerID))
			} else if card, ok := l[p.CardNumber]; !ok {
				log.Printf("WARN %v", fmt.Errorf("Card %v not initialised for controller %v", p.CardNumber, door.ControllerID))
			} else {
				card.Doors[door.Door] = true
			}
		}
	}

	return &acl, nil
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
