package system

import (
	"fmt"
	"sort"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func Events(uid, role string, start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)

	return sys.events.AsObjects(start, count, auth)
}

func AppendEvents(list types.EventsList) {
	deviceID := list.DeviceID
	recent := list.Events

	l := func(e uhppoted.Event) (string, string, string) {
		device := eventController(e)
		door := eventDoor(e)
		card := eventCard(e)

		return device, door, card
	}

	sys.events.Received(deviceID, recent, l)

	if len(recent) > 0 {
		if err := save(sys.events.file, sys.events.tag, &sys.events); err != nil {
			warn(err)
		}
	}
}

func eventController(e uhppoted.Event) string {
	name := ""

	if e.DeviceID != 0 {
		if oid := catalog.FindController(e.DeviceID); oid != "" {
			if v := catalog.GetV(oid, catalog.ControllerName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := sys.logs.Query("controller", fmt.Sprintf("%v", e.DeviceID), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		timestamp := time.Time(e.Timestamp)
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

func eventCard(e uhppoted.Event) string {
	name := ""

	if e.CardNumber != 0 {
		if oid, ok := catalog.Find(catalog.CardsOID, catalog.CardNumber, e.CardNumber); ok && oid != "" {
			oid = oid.Trim(catalog.CardNumber)
			if v := catalog.GetV(oid, catalog.CardName); v != nil {
				name = fmt.Sprintf("%v", v)
			}
		}

		edits := sys.logs.Query("card", fmt.Sprintf("%v", e.CardNumber), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		timestamp := time.Time(e.Timestamp)
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

func eventDoor(e uhppoted.Event) string {
	name := ""

	if e.DeviceID != 0 && e.Door >= 1 && e.Door <= 4 {
		if oid := catalog.FindController(e.DeviceID); oid != "" {
			var door interface{}

			switch e.Door {
			case 1:
				door = catalog.GetV(oid, catalog.ControllerDoor1)
			case 2:
				door = catalog.GetV(oid, catalog.ControllerDoor2)
			case 3:
				door = catalog.GetV(oid, catalog.ControllerDoor3)
			case 4:
				door = catalog.GetV(oid, catalog.ControllerDoor4)
			}

			if door != nil {
				if v := catalog.GetV(door.(catalog.OID), catalog.DoorName); v != nil {
					name = fmt.Sprintf("%v", v)
				}
			}
		}

		edits := sys.logs.Query("door", fmt.Sprintf("%v:%v", e.DeviceID, e.Door), "name")

		sort.SliceStable(edits, func(i, j int) bool {
			p := edits[i].Timestamp
			q := edits[j].Timestamp

			return q.Before(p)
		})

		timestamp := time.Time(e.Timestamp)
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
