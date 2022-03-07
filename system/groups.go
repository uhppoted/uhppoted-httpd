package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Groups(uid, role string) []catalog.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.groups.AsObjects(auth)

	return objects
}

func UpdateGroups(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.groups.Clone()

	for _, o := range objects {
		if updated, err := shadow.UpdateByOID(auth, o.OID, o.Value, dbc); err != nil {
			return nil, err
		} else {
			dbc.Stash(updated)
		}
	}

	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	if err := save(sys.groups.file, sys.groups.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.groups.Groups = shadow
	sys.updated()

	list := squoosh(dbc.Objects())

	return list, nil
}
