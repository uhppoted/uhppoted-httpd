package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateDoors(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.doors.doors.Clone()

	for _, o := range objects {
		if updated, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(updated)
		}
	}

	// ... validate
	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	for _, c := range sys.controllers.controllers.Controllers {
		for k, v := range c.Doors {
			if v != "" {
				if door, ok := shadow.Doors[catalog.OID(v)]; !ok {
					return nil, types.BadRequest(fmt.Errorf("Door %v not defined for controller %v", k, c), fmt.Errorf("controller %v: invalid door (%v)", c, k))

				} else if door.IsDeleted() {
					name := fmt.Sprintf("%v", door)

					if name == "" {
						return nil, types.BadRequest(fmt.Errorf("Deleting door in use by controller %v", c), fmt.Errorf("door %v: deleting door in use by controller %v", v, c))
					} else {
						return nil, types.BadRequest(fmt.Errorf("Deleting door %v in use by controller %v", door, c), fmt.Errorf("door %v: deleting door in use by controller %v", v, c))
					}
				}
			}
		}
	}

	// ... save
	if err := save(sys.doors.file, sys.doors.tag, &shadow); err != nil {
		return nil, err
	}

	sys.doors.doors = shadow
	dbc.Commit()
	sys.updated()

	list := squoosh(dbc.Objects())
	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: list,
	}, nil
}
