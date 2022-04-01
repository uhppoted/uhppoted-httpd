package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Controllers(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.controllers.AsObjects(auth)

	return objects
}

func UpdateControllers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.controllers.Controllers.Clone()

	for _, o := range updated {
		if objects, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	for _, oid := range deleted {
		if objects, err := shadow.DeleteByOID(auth, oid, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	if err := validate(&shadow); err != nil {
		return nil, err
	}

	if err := save(sys.controllers.file, sys.controllers.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	shadow.Committed()
	sys.controllers.Controllers = shadow
	sys.updated()

	list := squoosh(dbc.Objects())

	return list, nil
}

func validate(c *controllers.Controllers) error {
	if err := c.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	doors := map[schema.OID]string{}
	controllers := c.List()

	for _, r := range controllers {
		for _, v := range r.Doors {
			if v != "" {
				if _, ok := sys.doors.Door(schema.OID(v)); !ok {
					return types.BadRequest(
						fmt.Errorf("Invalid door ID"),
						fmt.Errorf("controller %v: invalid door ID (%v)", r.OIDx(), v))
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				d, _ := sys.doors.Door(v)
				return types.BadRequest(
					fmt.Errorf("%v door assigned to more than one controller", d.Name),
					fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OIDx()))
			}

			doors[v] = string(r.OIDx())
		}
	}

	return nil
}
