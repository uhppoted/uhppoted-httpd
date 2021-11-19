package system

import (
//	"github.com/uhppoted/uhppoted-httpd/auth"
//"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateUsers(uid, old, pwd string) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	//	if uid == "" || uid != auth.UID() {
	//		return nil, types.BadRequest(fmt.Errorf("Invalid user ID or password"), fmt.Errorf("update password: UID does not match session user"))
	//	}
	//
	//	if err := auth.Verify(uid, old); err != nil {
	//		return nil, types.BadRequest(fmt.Errorf("Invalid user ID or password"), fmt.Errorf("update password: invalid UID and/or PWD for %va, auth.UID()"))
	//	}

	return struct {
	}{}, nil
}
