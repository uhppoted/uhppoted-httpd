package system

import (
	"sync"
	"time"

	lib "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
)

func (s *system) synchronizeACL() error {
	year := time.Now().Year()
	from := lib.ToDate(year, time.January, 1)
	to := lib.ToDate(year, time.December, 31)
	controllers := s.controllers.AsIControllers()

	if acl, err := s.permissions(controllers); err != nil {
		return err
	} else {
		var wg sync.WaitGroup
		for _, c := range controllers {
			wg.Add(1)

			controller := c
			list := []lib.Card{}
			for _, card := range acl[controller.ID()] {
				list = append(list, card)
			}

			go func(v types.IController, cards []lib.Card) {
				defer wg.Done()

				for _, card := range cards {
					start := from
					end := to
					permissions := card.Doors

					if card.From != nil {
						start = *card.From
					}

					if card.To != nil {
						end = *card.To
					}

					s.interfaces.SetCard(controller, card.CardNumber, start, end, permissions)
				}
			}(controller, list)
		}

		wg.Wait()
	}

	return nil
}

func (s *system) compareACL() {
	controllers := s.controllers.AsIControllers()

	if acl, err := s.permissions(controllers); err != nil {
		warnf("%v", err)
	} else if diff, err := s.interfaces.CompareACL(controllers, acl); err != nil {
		warnf("%v", err)
	} else if diff == nil {
		warnf("Invalid ACL diff (%v)", diff)
	} else {
		cards := map[uint32]struct{}{}

		for _, v := range diff {
			for _, v := range v.Updated {
				cards[v.CardNumber] = struct{}{}
			}

			for _, v := range v.Added {
				cards[v.CardNumber] = struct{}{}
			}

			for _, v := range v.Deleted {
				cards[v.CardNumber] = struct{}{}
			}
		}

		list := []uint32{}
		for k, _ := range cards {
			list = append(list, k)
		}

		sys.cards.MarkIncorrect(list)
	}
}

// NTS: revoke all if card is nil because card number may have changed and the old
//      card will no longer have access
func (s *system) updateCardPermissions(controller types.IController, cardID uint32) {
	acl := map[uint8]uint8{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}

	year := time.Now().Year()
	from := lib.ToDate(year, time.January, 1)
	to := lib.ToDate(year, time.December, 31)

	if card := s.cards.Lookup(cardID); card != nil {
		from = card.From()
		to = card.To()

		// ... get base permissions from groups
		groups := card.Groups()
		doors := s.groups.Doors(groups...)

		for _, door := range doors {
			for _, d := range []uint8{1, 2, 3, 4} {
				doorID := d
				if oid, ok := controller.Door(d); ok && oid == door {
					acl[doorID] = 1
				}
			}
		}

		// ... updated base permissions with grules
		if sys.rules != nil {
			allowed, forbidden, err := sys.rules.Eval(*card, sys.doors)
			if err != nil {
				warn(err)
				return
			}

			for _, door := range allowed {
				for _, d := range []uint8{1, 2, 3, 4} {
					doorID := d
					if oid, ok := controller.Door(d); ok && oid == door.OID {
						acl[doorID] = 1
					}
				}
			}

			for _, door := range forbidden {
				for _, d := range []uint8{1, 2, 3, 4} {
					doorID := d
					if oid, ok := controller.Door(d); ok && oid == door.OID {
						acl[doorID] = 0
					}
				}
			}
		}
	}

	s.interfaces.SetCard(controller, cardID, from, to, acl)
}

func (s *system) permissions(controllers []types.IController) (acl.ACL, error) {
	cards := s.cards.List()
	groups := s.groups
	doors := s.doors

	// initialise empty ACL
	acl := make(acl.ACL)
	for _, b := range controllers {
		if v := b.ID(); v != 0 {
			acl[v] = map[uint32]lib.Card{}
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
		card := c.CardID
		membership := c.Groups()
		for _, g := range membership {
			if group, ok := groups.Group(g); ok {
				for d, allowed := range group.Doors {
					if door, ok := doors.Door(d); ok && allowed {
						device := catalog.GetDoorDeviceID(door.OID)
						doorID := catalog.GetDoorDeviceDoor(door.OID)

						grant(card, device, doorID)
					}
				}
			}
		}
	}

	// ... post-process ACL with rules
	if sys.rules != nil {
		for _, c := range cards {
			card := c.CardID
			allowed, forbidden, err := sys.rules.Eval(c, sys.doors)
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

	return acl, nil
}
