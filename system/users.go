package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Users(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.users.AsObjects(auth)

	return objects
}

func UpdateUsers(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Users.Clone()

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

func User(uid string) (auth.IUser, bool) {
	return sys.users.User(uid)
}

func SetPassword(uid, pwd string) error {
	sys.Lock()
	defer sys.Unlock()

	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Users.Clone()

	if updated, err := shadow.SetPassword(uid, pwd, dbc); err != nil {
		return err
	} else {
		dbc.Stash(updated)
	}

	if err := shadow.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	if err := save(sys.users.file, sys.users.tag, &shadow); err != nil {
		return err
	}

	dbc.Commit()
	sys.users.Users = shadow
	sys.updated()

	return nil
}
