package system

import (
	"bytes"
	"fmt"
	"log"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/acl"
)

func UpdateACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else {
		sys.controllers.controllers.UpdateACL(acl)
	}
}

func CompareACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else if err := sys.controllers.controllers.CompareACL(acl); err != nil {
		warn(err)
	}
}

func permissions() (acl.ACL, error) {
	cards := sys.cards.cards.Cards
	groups := sys.groups.groups.Groups
	doors := sys.doors.doors.Doors

	// initialise empty ACL

	acl := make(acl.ACL)

	for _, b := range sys.controllers.controllers.Controllers {
		if v := b.DeviceID(); v != 0 {
			acl[v] = map[uint32]core.Card{}
		}
	}

	for _, l := range acl {
		for _, c := range cards {
			if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
				card := uint32(*c.Card)
				from := core.Date(*c.From)
				to := core.Date(*c.To)

				l[card] = core.Card{
					CardNumber: card,
					From:       &from,
					To:         &to,
					Doors:      map[uint8]int{1: 0, 2: 0, 3: 0, 4: 0},
				}
			}
		}
	}

	// ... populate ACL from cards + groups + doors

	grant := func(card uint32, device uint32, door uint8) {
		if card > 0 && device > 0 && door >= 1 && door <= 4 {
			if _, ok := acl[device]; ok {
				if _, ok := acl[device][card]; ok {
					if _, ok := acl[device][card].Doors[door]; ok {
						acl[device][card].Doors[door] = 1
					}
				}
			}
		}
	}

	revoke := func(card uint32, device uint32, door uint8) {
		if card > 0 && device > 0 && door >= 1 && door <= 4 {
			if _, ok := acl[device]; ok {
				if _, ok := acl[device][card]; ok {
					if _, ok := acl[device][card].Doors[door]; ok {
						acl[device][card].Doors[door] = 0
					}
				}
			}
		}
	}

	for _, c := range cards {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			for g, member := range c.Groups {
				if group, ok := groups[catalog.OID(g)]; ok && member {
					for d, allowed := range group.Doors {
						if door, ok := doors[d]; ok && allowed {
							card := uint32(*c.Card)
							device := catalog.GetDoorDeviceID(door.OID)
							doorID := catalog.GetDoorDeviceDoor(door.OID)

							grant(card, device, doorID)
						}
					}
				}
			}
		}
	}

	// ... post-process ACL with rules

	if sys.rules != nil {
		for _, c := range cards {
			allowed, forbidden, err := sys.rules.Eval(*c, sys.groups.groups, sys.doors.doors)
			if err != nil {
				return nil, err
			}

			for _, door := range allowed {
				card := uint32(*c.Card)
				device := catalog.GetDoorDeviceID(door.OID)
				doorID := catalog.GetDoorDeviceDoor(door.OID)
				grant(card, device, doorID)
			}

			for _, door := range forbidden {
				card := uint32(*c.Card)
				device := catalog.GetDoorDeviceID(door.OID)
				doorID := catalog.GetDoorDeviceDoor(door.OID)
				revoke(card, device, doorID)
			}
		}
	}

	// ... 'k, done

	var b bytes.Buffer

	acl.Print(&b)
	log.Printf("INFO %v", fmt.Sprintf("ACL\n%s", string(b.Bytes())))

	return acl, nil
}
