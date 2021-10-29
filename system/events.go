package system

import (
	"fmt"
	"os/exec"
	"sort"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func (f *callback) Append(deviceID uint32, recent []uhppoted.Event) {
	l := func(e uhppoted.Event) (string, string, string) {
		device := eventController(e)
		door := ""
		card := eventCard(e)

		if c := sys.controllers.Find(e.DeviceID); c != nil {
			if d, ok := c.Door(e.Door); ok {
				door = d
			}
		}

		return device, door, card
	}

	sys.events.Received(deviceID, recent, l)

	if len(recent) > 0 {
		sys.events.Save()
	}
}

func eventController(e uhppoted.Event) string {
	name := ""

	if e.DeviceID != 0 {
		if oid := catalog.FindController(e.DeviceID); oid != "" {
			if v, _ := catalog.GetV(oid.Append(catalog.ControllerName)); v != nil {
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
			if v.Timestamp.Before(timestamp) {
				switch {
				case v.After != "":
					name = v.After
					break
				case v.Before != "":
					name = v.Before
					break
				}
			}
		}
	}

	return name
}

func eventCard(e uhppoted.Event) string {
	name := ""

	if e.CardNumber == 0 {
		return name
	}

	if oid, ok := catalog.Find(catalog.CardsOID, catalog.CardNumber, e.CardNumber); ok && oid != "" {
		oid = oid.Trim(catalog.CardNumber)
		if v, _ := catalog.GetV(oid.Append(catalog.CardName)); v != nil {
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
		if v.Timestamp.Before(timestamp) {
			switch {
			case v.After != "":
				name = v.After
				break
			case v.Before != "":
				name = v.Before
				break
			}
		}
	}

	return name
}
