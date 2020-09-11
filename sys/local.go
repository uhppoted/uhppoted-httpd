package system

import (
	"fmt"
	"log"

	uhppote "github.com/uhppoted/uhppote-core/uhppote"
	api "github.com/uhppoted/uhppoted-api/acl"
	acl "github.com/uhppoted/uhppoted-httpd/acl"
)

type Local struct {
	u uhppote.UHPPOTE

	devices []*uhppote.Device
}

func (l *Local) Update(permissions acl.ACL) {
	log.Printf("Updating ACL")

	fmt.Printf(">> PERMISSIONS\n%v\n", permissions)

	table, err := permissionsToTable(permissions)
	if err != nil {
		warn(err)
		return
	}

	fmt.Printf(">> TABLE\n%v\n", table)

	list, _, err := api.ParseTable(table, l.devices, true)
	if err != nil {
		warn(err)
		return
	}

	fmt.Printf(">> ACL\n%v\n", list)

	rpt, err := api.PutACL(&l.u, *list, false)
	if err != nil {
		warn(err)
		return
	}

	log.Printf("ACL updated")
	log.Printf("%v", rpt)
}
