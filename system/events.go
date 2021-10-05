package system

import (
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func (f *callback) Append(deviceID uint32, recent []uhppoted.Event) {
	lookup := func(e uhppoted.Event) (string, string, string) {
		device := ""
		door := ""
		card := ""

		if c := sys.controllers.Lookup(e.DeviceID); c != nil {
			if c.Name != nil {
				device = string(*c.Name)
			}

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

	sys.events.Received(deviceID, recent, lookup)

	//		if len(recent) > 0 {
	//			sys.events.Save()
	//		}
}
