package system

import (
	"fmt"
	"sort"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type histxxx struct {
}

func (h histxxx) lookupController(timestamp time.Time, deviceID uint32) string {
	name := ""

	if deviceID != 0 {
		if oid := catalog.FindController(deviceID); oid != "" {
			if v := catalog.GetV(oid, schema.ControllerName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := sys.logs.Query("controller", fmt.Sprintf("%v", deviceID), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		for _, v := range edits {
			if v.Timestamp.After(timestamp) {
				switch {
				case v.Before != "":
					name = v.Before
					break
				case v.After != "":
					name = v.After
					break
				}
			}
		}
	}

	return name
}

func (h histxxx) lookupCard(timestamp time.Time, card uint32) string {
	name := ""

	if card != 0 {
		if oid, ok := catalog.Find(schema.CardsOID, schema.CardNumber, card); ok && oid != "" {
			oid = oid.Trim(schema.CardNumber)
			if v := catalog.GetV(oid, schema.CardName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := sys.logs.Query("card", fmt.Sprintf("%v", card), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		for _, v := range edits {
			if v.Timestamp.After(timestamp) {
				switch {
				case v.Before != "":
					name = v.Before
					break
				case v.After != "":
					name = v.After
					break
				}
			}
		}
	}

	return name
}

func (h histxxx) lookupDoor(timestamp time.Time, deviceID uint32, door uint8) string {
	name := ""

	if deviceID != 0 && door >= 1 && door <= 4 {
		if oid := catalog.FindController(deviceID); oid != "" {
			var dOID interface{}

			switch door {
			case 1:
				dOID = catalog.GetV(oid, schema.ControllerDoor1)
			case 2:
				dOID = catalog.GetV(oid, schema.ControllerDoor2)
			case 3:
				dOID = catalog.GetV(oid, schema.ControllerDoor3)

			case 4:
				dOID = catalog.GetV(oid, schema.ControllerDoor4)
			}

			if dOID != nil {
				if v := catalog.GetV(dOID.(schema.OID), schema.DoorName); v != nil {
					name = fmt.Sprintf("%v", v)
				}
			}
		}

		edits := sys.logs.Query("door", fmt.Sprintf("%v:%v", deviceID, door), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		for _, v := range edits {
			if v.Timestamp.After(timestamp) {
				switch {
				case v.Before != "":
					name = v.Before
					break
				case v.After != "":
					name = v.After
					break
				}
			}
		}
	}

	return name
}
