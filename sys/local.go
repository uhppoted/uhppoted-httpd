package system

import (
	"fmt"
	"log"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-api/acl"
)

type Local struct {
	u uhppote.UHPPOTE

	devices []*uhppote.Device
}

func (l *Local) Update(permissions []Permissions) {
	log.Printf("Updating ACL")

	fmt.Printf(">> PERMISSIONS\n%v\n", permissions)

	table, err := consolidate(permissions)
	if err != nil {
		warn(err)
		return
	}

	fmt.Printf(">> TABLE\n%v\n", table)

	list, _, err := acl.ParseTable(table, l.devices, true)
	if err != nil {
		warn(err)
		return
	}

	fmt.Printf(">> ACL\n%v\n", list)

	rpt, err := acl.PutACL(&l.u, *list, false)
	if err != nil {
		warn(err)
		return
	}

	log.Printf("ACL updated")
	log.Printf("%v", rpt)
}
