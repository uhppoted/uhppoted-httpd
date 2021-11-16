package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateControllers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.controllers.Clone()

	for _, o := range objects {
		if updated, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(updated)
		}
	}

	if err := save(&shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.controllers = shadow
	sys.updated()

	list := squoosh(dbc.Objects())
	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: list,
	}, nil
}

func save(c *controllers.ControllerSet) error {
	if err := validate(c); err != nil {
		return err
	}

	return c.Save()
}

func validate(c *controllers.ControllerSet) error {
	if err := c.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	doors := map[catalog.OID]string{}

	for _, r := range c.Controllers {
		for _, v := range r.Doors {
			if v != "" {
				if _, ok := sys.doors.Doors[catalog.OID(v)]; !ok {
					return types.BadRequest(fmt.Errorf("Invalid door ID"), fmt.Errorf("controller %v: invalid door ID (%v)", r.OID, v))
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				return types.BadRequest(fmt.Errorf("%v door assigned to more than one controller", sys.doors.Doors[catalog.OID(v)].Name), fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID))
			}

			doors[v] = string(r.OID)
		}
	}

	return nil
}
