package system

import (
	"bytes"
	"fmt"
	"log"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/acl"
)

const DoorControllerID = catalog.DoorControllerID
const DoorControllerDoor = catalog.DoorControllerDoor

func UpdateACL(rules cards.IRules) {
	if acl, err := permissions(rules); err != nil {
		warn(err)
	} else {
		sys.controllers.UpdateACL(acl)
	}
}

func CompareACL(rules cards.IRules) {
	if acl, err := permissions(rules); err != nil {
		warn(err)
	} else if err := sys.controllers.Compare(acl); err != nil {
		warn(err)
	}
}

func permissions(rules cards.IRules) (acl.ACL, error) {
	lCards := sys.cards.List()
	lGroups := sys.groups.List()
	lDoors := sys.doors.List()

	// initialise empty ACL
	acl := make(acl.ACL)

	for _, b := range sys.controllers.Controllers {
		if b.DeviceID != nil && *b.DeviceID > 0 {
			acl[*b.DeviceID] = map[uint32]core.Card{}
		}
	}

	for _, l := range acl {
		for _, c := range lCards {
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
	for _, c := range lCards {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			for k, ok := range c.Groups {
				if ok {
					for _, g := range lGroups {
						if g.OID == catalog.OID(k) {
							for d, ok := range g.Doors {
								if ok {
									for _, dd := range lDoors {
										if dd.OID == d {
											if v, _ := catalog.GetV(dd.OID.Append(DoorControllerID)); v != nil {
												if w, _ := catalog.GetV(dd.OID.Append(DoorControllerDoor)); w != nil {
													card := uint32(*c.Card)
													device := *(v.(*uint32))
													door := w.(uint8)

													if _, ok := acl[device]; ok {
														if _, ok := acl[device][card]; ok {
															if _, ok := acl[device][card].Doors[door]; ok {
																acl[device][card].Doors[door] = 1
															}
														}
													}

												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			// var doors = []string{}
			//			var err error
			//
			//			if rules != nil {
			//				doors, err = rules.Eval(*c)
			//				if err != nil {
			//					return nil, err
			//				}
			//			}
			//
			//			permission := types.Permissions{
			//				CardNumber: uint32(*c.Card),
			//				From:       *c.From,
			//				To:         *c.To,
			//				Doors:      doors,
			//			}
			//
			//			list = append(list, permission)
		}
	}

	var b bytes.Buffer

	acl.Print(&b)
	log.Printf("INFO %v", fmt.Sprintf("ACL\n%s", string(b.Bytes())))

	return acl, nil
}

//func consolidate(list []types.Permissions) (*acl.ACL, error) {
//	// initialise empty ACL
//	acl := make(acl.ACL)
//
//	for _, c := range sys.controllers.Controllers {
//		if c.DeviceID != nil && *c.DeviceID > 0 {
//			acl[*c.DeviceID] = map[uint32]core.Card{}
//		}
//	}
//
//	// create ACL with all cards on all controllers
//	for _, p := range list {
//		for _, l := range acl {
//			if _, ok := l[p.CardNumber]; !ok {
//				from := core.Date(p.From)
//				to := core.Date(p.To)
//
//				l[p.CardNumber] = core.Card{
//					CardNumber: p.CardNumber,
//					From:       &from,
//					To:         &to,
//					Doors:      map[uint8]int{1: 0, 2: 0, 3: 0, 4: 0},
//				}
//			}
//		}
//	}
//
//	// update ACL cards from permissions
//
//	for _, p := range list {
//	loop:
//		for _, d := range p.Doors {
//			door, ok := sys.doors.Find(d)
//			if !ok {
//				log.Printf("WARN %v", fmt.Errorf("consolidate: undefined door '%v' for card %v", d, p.CardNumber))
//				continue
//			}
//
//			cid := lookup(door.OID.Append(catalog.DoorControllerOID)) // controller OID
//			did := uint8(0)                                           // controller door
//			if v := lookup(door.OID.Append(catalog.DoorControllerDoor)); v != nil {
//				if w, ok := v.(uint8); ok {
//					did = w
//				}
//			}
//
//			if cid != nil && did >= 1 && did <= 4 {
//				for _, c := range sys.controllers.Controllers {
//					if c.OID == cid {
//						if c.DeviceID != nil && *c.DeviceID > 0 {
//							if l, ok := acl[*c.DeviceID]; ok {
//								if card, ok := l[p.CardNumber]; !ok {
//									log.Printf("WARN %v", fmt.Errorf("consolidate: card %v not initialised for controller %v", p.CardNumber, *c.DeviceID))
//								} else {
//									card.Doors[did] = 1
//								}
//							}
//						}
//
//						continue loop
//					}
//				}
//			}
//
//			log.Printf("WARN %v", fmt.Errorf("consolidate: card %v, door %v - no controller assigned", p.CardNumber, door))
//		}
//	}
//
//	var b bytes.Buffer
//
//	acl.Print(&b)
//	log.Printf("INFO %v", fmt.Sprintf("ACL\n%s", string(b.Bytes())))
//
//	return &acl, nil
//}

func lookup(oid string) interface{} {
	v, _ := catalog.GetV(oid)

	return v
}
