package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
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

	updated, _, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.interfaces.Interfaces.Clone()

	for _, o := range updated {
		if objects, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(objects)
		}
	}

	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	if err := save(sys.interfaces.file, sys.interfaces.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	shadow.Committed()

	sys.interfaces.Interfaces = shadow
	sys.updated()

	list := squoosh(dbc.Objects())

	return list, nil
}
