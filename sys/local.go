package system

import (
	"fmt"
	"log"

	"github.com/uhppoted/uhppote-core/uhppote"
	//	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Local struct {
	u uhppote.UHPPOTE

	devices []*uhppote.Device
}

func (l *Local) Update(permissions []types.Permissions) {
	log.Printf("Updating ACL")

	acl, err := consolidate(permissions)
	if err != nil {
		warn(err)
		return
	} else if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
		return
	}

	fmt.Printf(">> >> ACL\n")
	for k, v := range *acl {
		fmt.Printf(">> >> %v\n", k)
		for _, w := range v {
			fmt.Printf("         %v\n", w)
		}
	}

	//rpt, err := acl.PutACL(&l.u, *list, false)
	//if err != nil {
	//	warn(err)
	//	return
	//}

	log.Printf("ACL updated")
	//log.Printf("%v", rpt)
}
