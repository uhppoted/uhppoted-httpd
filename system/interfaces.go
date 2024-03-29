package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

func Interfaces(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.interfaces.AsObjects(auth)

	return objects
}

func UpdateInterfaces(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	created, updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.interfaces.Clone()

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

	if err := shadow.Validate(); err != nil {
		return nil, err
	}

	if err := save(TagInterfaces, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.interfaces = shadow
	})

	return dbc.Objects(), nil
}
