package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

func Doors(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.doors.AsObjects(auth)

	return objects
}

func UpdateDoors(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	created, updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.doors.Clone()

	for _, o := range created {
		if objects, err := shadow.Create(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	for _, o := range updated {
		if objects, err := shadow.Update(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	for _, oid := range deleted {
		if objects, err := shadow.Delete(auth, oid, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	// ... validate
	if err := shadow.Validate(); err != nil {
		return nil, err
	}

	controllers := sys.controllers.List()
	for _, c := range controllers {
		for k, v := range c.Doors() {
			if v != "" {
				if door, ok := shadow.Door(v); !ok {
					return nil, fmt.Errorf("door %v not defined for controller %v", k, c)

				} else if door.IsDeleted() {
					name := fmt.Sprintf("%v", door)

					if name == "" {
						return nil, fmt.Errorf("deleting door in use by controller %v", c)
					} else {
						return nil, fmt.Errorf("deleting door %v in use by controller %v", door, c)
					}
				}
			}
		}
	}

	// ... save
	if err := save(TagDoors, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.doors = shadow
	})

	return dbc.Objects(), nil
}
