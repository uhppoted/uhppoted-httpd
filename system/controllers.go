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

func UpdateControllers(m map[string]interface{}, a *auth.Authorizator) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	created, updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.controllers.Clone()

	for _, o := range created {
		if objects, err := shadow.Create(a, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	for _, o := range updated {
		if objects, err := shadow.Update(a, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	for _, oid := range deleted {
		if objects, err := shadow.Delete(a, oid, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	if err := validate(&shadow); err != nil {
		return nil, err
	}

	if err := save(TagControllers, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.controllers = shadow
	})

	return dbc.Objects(), nil
}

func validate(cc *controllers.Controllers) error {
	if err := cc.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	doors := map[schema.OID]string{}
	controllers := cc.List()

	for _, r := range controllers {
		for _, v := range r.Doors() {
			if v != "" {
				if _, ok := sys.doors.Door(schema.OID(v)); !ok {
					return types.BadRequest(
						fmt.Errorf("Invalid door ID"),
						fmt.Errorf("controller %v: invalid door ID (%v)", r.OID, v))
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				d, _ := sys.doors.Door(v)
				return types.BadRequest(
					fmt.Errorf("%v door assigned to more than one controller", d),
					fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID))
			}

			doors[v] = string(r.OID)
		}
	}

	return nil
}
