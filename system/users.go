package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Users(auth auth.OpAuth) interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.users.AsObjects(auth)
}

func UpdateUsers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Users.Clone()

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

	if err := save(sys.users.file, sys.users.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.users.Users = shadow
	sys.updated()

	list := squoosh(dbc.Objects())

	return list, nil
}
