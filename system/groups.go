package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateGroups(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	list := struct {
		Objects []interface{} `json:"objects,omitempty"`
	}{}

	shadow := sys.groups.Clone()

	for _, object := range objects {
		if updated, err := shadow.UpdateByOID(auth, object.OID, object.Value); err != nil {
			return nil, err
		} else if updated != nil {
			list.Objects = append(list.Objects, updated...)
		}
	}

	// ... validate
	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	// for _, c := range sys.controllers.Controllers {
	//     for k, v := range c.Doors {
	//         if v != "" {
	//             if door, ok := shadow.Doors[v]; !ok {
	//                 return nil, types.BadRequest(fmt.Errorf("Door %v not defined for controller %v", k, c), fmt.Errorf("controller %v: invalid door (%v)", c, k))
	//
	//             } else if door.IsDeleted() {
	//                 name := fmt.Sprintf("%v", door)
	//
	//                 if name == "" {
	//                     return nil, types.BadRequest(fmt.Errorf("Deleting door in use by controller %v", c), fmt.Errorf("door %v: deleting door in use by controller %v", v, c))
	//                 } else {
	//                     return nil, types.BadRequest(fmt.Errorf("Deleting door %v in use by controller %v", door, c), fmt.Errorf("door %v: deleting door in use by controller %v", v, c))
	//                 }
	//             }
	//         }
	//     }
	// }

	// ... save
	if err := shadow.Save(); err != nil {
		return nil, err
	}

	sys.groups = *shadow
	sys.updated()

	return list, nil
}
