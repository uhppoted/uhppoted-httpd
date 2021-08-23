package system

import (
	"bytes"
	"fmt"
	"log"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
)

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
			door, ok := sys.doors.Find(d)
			if !ok {
				log.Printf("WARN %v", fmt.Errorf("consolidate: undefined door '%v' for card %v", d, p.CardNumber))
				continue
			}

			cid := lookup(door.OID + catalog.DoorControllerOID) // controller OID
			did := uint8(0)                                     // controller door
			if v := lookup(door.OID + catalog.DoorControllerDoor); v != nil {
				if w, ok := v.(uint8); ok {
					did = w
				}
			}

			if cid != nil && did >= 1 && did <= 4 {
				for _, c := range sys.controllers.Controllers {
					if c.OID == cid {
						if c.DeviceID != nil && *c.DeviceID > 0 {
							if l, ok := acl[*c.DeviceID]; ok {
								if card, ok := l[p.CardNumber]; !ok {
									log.Printf("WARN %v", fmt.Errorf("consolidate: card %v not initialised for controller %v", p.CardNumber, *c.DeviceID))
								} else {
									card.Doors[did] = 1
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

func lookup(oid string) interface{} {
	v, _ := catalog.GetV(oid)

	return v
}
