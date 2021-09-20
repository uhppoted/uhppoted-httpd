package system

import (
	"bytes"
	"fmt"
	"log"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/acl"
)

const DoorControllerID = catalog.DoorControllerID
const DoorControllerDoor = catalog.DoorControllerDoor

func UpdateACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else {
		sys.controllers.UpdateACL(acl)
	}
}

func CompareACL() {
	if acl, err := permissions(); err != nil {
		warn(err)
	} else if err := sys.controllers.Compare(acl); err != nil {
		warn(err)
	}
}

func permissions() (acl.ACL, error) {
	cards := sys.cards.List()
	groups := sys.groups.List()
	doors := sys.doors.List()

	// initialise empty ACL
	acl := make(acl.ACL)

	for _, b := range sys.controllers.Controllers {
		if b.DeviceID != nil && *b.DeviceID > 0 {
			acl[*b.DeviceID] = map[uint32]core.Card{}
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
	for _, c := range cards {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			for k, ok := range c.Groups {
				if ok {
					for _, g := range groups {
						if g.OID == catalog.OID(k) {
							for d, ok := range g.Doors {
								if ok {
									for _, dd := range doors {
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
