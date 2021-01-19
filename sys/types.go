package system

import (
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type datetime struct {
	DateTime *types.DateTime
	TimeZone *time.Location
	Status   status
}

type ip struct {
	IP     *address
	Status status
}

// func (a *ip) String() string{
// 	if a != nil {
// 		return a.IP.String()
// 	}
//
// 	return ""
// }

type records uint32

func (r *records) String() string {
	if r != nil {
		return fmt.Sprintf("%v", uint32(*r))
	}

	return ""
}
