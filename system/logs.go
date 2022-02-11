package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

func Logs(uid, role string, start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)

	return sys.logs.AsObjects(start, count, auth)
}
