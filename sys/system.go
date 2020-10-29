package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type System struct {
	Doors map[string]types.Door `json:"doors"`
	Local []Local               `json:"local"`
}

var sys = System{
	Doors: map[string]types.Door{},
	Local: []Local{},
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
