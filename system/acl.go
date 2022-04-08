package system

import (
	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/acl"
)

func UpdateACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else {
		sys.controllers.UpdateACL(sys.interfaces.Interfaces, acl)
	}
}

func CompareACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else if err := sys.controllers.CompareACL(sys.interfaces.Interfaces, acl); err != nil {
		warn(err)
	}
}

func permissions() (acl.ACL, error) {
	cards := sys.cards.List()
	groups := sys.groups.Groups
	doors := sys.doors.Doors
	controllers := sys.controllers.List()

	// initialise empty ACL
	acl := make(acl.ACL)
	for _, b := range controllers {
		if v := b.ID(); v != 0 {
			acl[v] = map[uint32]types.Card{}
		}
	}

	for _, l := range acl {
		for _, c := range cards {
			if card, ok := c.AsAclCard(); ok {
				l[card.CardNumber] = card
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
		if card, ok := c.AsAclCard(); ok {
			for g, member := range c.Groups() {
				if group, ok := groups.Group(g); ok && member {
					for d, allowed := range group.Doors {
						if door, ok := doors.Door(d); ok && allowed {
							device := catalog.GetDoorDeviceID(door.OID)
							doorID := catalog.GetDoorDeviceDoor(door.OID)

							grant(card.CardNumber, device, doorID)
						}
					}
				}
			}
		}
	}

	// ... post-process ACL with rules

	if sys.rules != nil {
		for _, c := range cards {
			card := c.CardNumber()
			allowed, forbidden, err := sys.rules.Eval(c, sys.doors.Doors)
			if err != nil {
				return nil, err
			}

			for _, door := range allowed {
				device := catalog.GetDoorDeviceID(door.OID)
				doorID := catalog.GetDoorDeviceDoor(door.OID)
				grant(card, device, doorID)
			}

			for _, door := range forbidden {
				device := catalog.GetDoorDeviceID(door.OID)
				doorID := catalog.GetDoorDeviceDoor(door.OID)
				revoke(card, device, doorID)
			}
		}
	}

	// ... 'k, done

	// var b bytes.Buffer
	//
	// acl.Print(&b)
	// log.Printf("INFO %v", fmt.Sprintf("ACL\n%s", string(b.Bytes())))

	return acl, nil
}
