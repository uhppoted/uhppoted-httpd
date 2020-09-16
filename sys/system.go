package system

import (
	"log"
	"strings"

	uhppote "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type System struct {
	Doors map[string]types.Door
}

var sys = System{
	Doors: map[string]types.Door{
		"D11": types.Door{
			ID:           "D11",
			ControllerID: 405419896,
			Door:         1,
			Name:         "Great Hall",
		},
		"D12": types.Door{
			ID:           "D12",
			ControllerID: 405419896,
			Door:         2,
			Name:         "Kitchen",
		},
		"D13": types.Door{
			ID:           "D13",
			ControllerID: 405419896,
			Door:         3,
			Name:         "Dungeon",
		},
		"D14": types.Door{
			ID:           "D14",
			ControllerID: 405419896,
			Door:         4,
			Name:         "Hogsmeade",
		},

		"D21": types.Door{
			ID:           "D21",
			ControllerID: 303986753,
			Door:         1,
			Name:         "Gryffindor",
		},
		"D22": types.Door{
			ID:           "D22",
			ControllerID: 303986753,
			Door:         2,
			Name:         "Hufflepuff",
		},
		"D23": types.Door{
			ID:           "D23",
			ControllerID: 303986753,
			Door:         3,
			Name:         "Ravenclaw",
		},
		"D24": types.Door{
			ID:           "D24",
			ControllerID: 303986753,
			Door:         4,
			Name:         "Slytherin",
		},
	},
}

var local = Local{}

func Update(permissions []types.Permissions) {
	local.Update(permissions)
}

func consolidate(list []types.Permissions) (*acl.ACLN, error) {
	acl := make(acl.ACLN)

	for _, d := range sys.Doors {
		if _, ok := acl[d.ControllerID]; !ok {
			acl[d.ControllerID] = make(map[uint32]uhppote.CardN)
		}
	}

	for _, p := range list {
		for k, l := range acl {
			card, ok := l[p.CardNumber]
			if !ok {
				from := uhppote.Date(p.From)
				to := uhppote.Date(p.To)

				card = uhppote.CardN{
					CardNumber: p.CardNumber,
					From:       &from,
					To:         &to,
					Doors:      map[uint8]bool{1: false, 2: false, 3: false, 4: false},
				}

				l[p.CardNumber] = card
			}

			for _, d := range p.Doors {
				if door, ok := sys.Doors[d]; ok {
					if door.ControllerID == k {
						card.Doors[door.Door] = true
					}
				}
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
