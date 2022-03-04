package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

func Logs(uid, role string, start, count int) []catalog.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)

	return sys.logs.AsObjects(start, count, auth).AsArray()
}
