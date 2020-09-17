package system

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/uhppoted/uhppote-core/uhppote"
	uhppoted "github.com/uhppoted/uhppoted-api/acl"
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
	}

	if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
		return
	}

	rpt, err := uhppoted.PutACLN(&l.u, *acl, false)
	if err != nil {
		warn(err)
		return
	}

	keys := []uint32{}
	for k, _ := range rpt {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "ACL updated\n")

	for _, k := range keys {
		v := rpt[k]
		fmt.Fprintf(&msg, "                    %v", k)
		fmt.Fprintf(&msg, " unchanged:%-3v", len(v.Unchanged))
		fmt.Fprintf(&msg, " updated:%-3v", len(v.Updated))
		fmt.Fprintf(&msg, " added:%-3v", len(v.Added))
		fmt.Fprintf(&msg, " deleted:%-3v", len(v.Deleted))
		fmt.Fprintf(&msg, " failed:%-3v", len(v.Failed))
		fmt.Fprintf(&msg, " errored:%-3v", len(v.Errored))
		fmt.Fprintln(&msg)
	}

	log.Printf("%v", string(msg.Bytes()))
}
