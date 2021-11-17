package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

func UpdateUsers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	//	fmt.Printf(">>>>>>>>> UID:  %v\n", m["uid"])
	//	fmt.Printf(">>>>>>>>> OLD:  %v\n", m["old"])
	//	fmt.Printf(">>>>>>>>> PWD:  %v\n", m["pwd"])
	//	fmt.Printf(">>>>>>>>> PWD2: %v\n", m["pwd2"])

	return struct {
	}{}, nil
}
