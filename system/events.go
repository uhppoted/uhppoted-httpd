package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func (f *callback) Append(deviceID uint32, recent []uhppoted.Event) {
	l := func(e uhppoted.Event) (string, string, string) {
		device := lookup(e)
		door := ""
		card := ""

		if c := sys.controllers.Find(e.DeviceID); c != nil {
			if d, ok := c.Door(e.Door); ok {
				door = d
			}
		}

		if c := sys.cards.Lookup(e.CardNumber); c != nil {
			if c.Name != nil {
				card = c.GetName()
			}
		}

		return device, door, card
	}

	sys.events.Received(deviceID, recent, l)

	if len(recent) > 0 {
		sys.events.Save()
	}
}

func lookup(e uhppoted.Event) string {
	name := ""

	if oid := catalog.FindController(e.DeviceID); oid != "" {
		if v, _ := catalog.GetV(oid.Append(catalog.ControllerName)); v != nil {
			name = fmt.Sprintf("%v", v)
		}
	}

	return name
}
