package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

func Logs(start, count int, auth auth.OpAuth) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.logs.AsObjects(start, count, auth)
}
