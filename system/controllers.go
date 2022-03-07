package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Controllers(uid, role string) []catalog.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.controllers.AsObjects(auth)

	return objects
}

func UpdateControllers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.controllers.Controllers.Clone()

	for _, o := range objects {
		if updated, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(updated)
		}
	}

	if err := validate(&shadow); err != nil {
		return nil, err
	}

	if err := save(sys.controllers.file, sys.controllers.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.controllers.Controllers = shadow
	sys.updated()

	list := squoosh(dbc.Objects())

	return list, nil
}

func validate(c *controllers.Controllers) error {
	if err := c.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	doors := map[catalog.OID]string{}
	controllers := c.List()

	for _, r := range controllers {
		for _, v := range r.Doors {
			if v != "" {
				if _, ok := sys.doors.Door(catalog.OID(v)); !ok {
					return types.BadRequest(
						fmt.Errorf("Invalid door ID"),
						fmt.Errorf("controller %v: invalid door ID (%v)", r.OID(), v))
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				d, _ := sys.doors.Door(v)
				return types.BadRequest(
					fmt.Errorf("%v door assigned to more than one controller", d.Name),
					fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID()))
			}

			doors[v] = string(r.OID())
		}
	}

	return nil
}
