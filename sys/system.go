package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type System struct {
	Doors map[string]types.Door `json:"doors"`
}

var sys = System{
	Doors: map[string]types.Door{},
	//	Doors: map[string]types.Door{
	//		"D11": types.Door{
	//			ID:           "D11",
	//			ControllerID: 405419896,
	//			Door:         1,
	//			Name:         "Great Hall",
	//		},
	//		"D12": types.Door{
	//			ID:           "D12",
	//			ControllerID: 405419896,
	//			Door:         2,
	//			Name:         "Kitchen",
	//		},
	//		"D13": types.Door{
	//			ID:           "D13",
	//			ControllerID: 405419896,
	//			Door:         3,
	//			Name:         "Dungeon",
	//		},
	//		"D14": types.Door{
	//			ID:           "D14",
	//			ControllerID: 405419896,
	//			Door:         4,
	//			Name:         "Hogsmeade",
	//		},
	//
	//		"D21": types.Door{
	//			ID:           "D21",
	//			ControllerID: 303986753,
	//			Door:         1,
	//			Name:         "Gryffindor",
	//		},
	//		"D22": types.Door{
	//			ID:           "D22",
	//			ControllerID: 303986753,
	//			Door:         2,
	//			Name:         "Hufflepuff",
	//		},
	//		"D23": types.Door{
	//			ID:           "D23",
	//			ControllerID: 303986753,
	//			Door:         3,
	//			Name:         "Ravenclaw",
	//		},
	//		"D24": types.Door{
	//			ID:           "D24",
	//			ControllerID: 303986753,
	//			Door:         4,
	//			Name:         "Slytherin",
	//		},
	//	},
}

var local = Local{
	u: uhppote.UHPPOTE{
		BindAddress:      resolve("192.168.1.100:0"),
		BroadcastAddress: resolve("192.168.1.255:60000"),
		ListenAddress:    resolve("192.168.1.100:60001"),
		Devices:          map[uint32]*uhppote.Device{},
		Debug:            false,
	},

	devices: []*uhppote.Device{},
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
	local.Update(permissions)
}

func consolidate(list []types.Permissions) (*acl.ACLN, error) {
	// initialise empty ACL
	acl := make(acl.ACLN)

	for _, d := range sys.Doors {
		if _, ok := acl[d.ControllerID]; !ok {
			acl[d.ControllerID] = make(map[uint32]core.CardN)
		}
	}

	// create ACL with all cards on all controllers
	for _, p := range list {
		for _, l := range acl {
			if _, ok := l[p.CardNumber]; !ok {
				from := core.Date(p.From)
				to := core.Date(p.To)

				l[p.CardNumber] = core.CardN{
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
