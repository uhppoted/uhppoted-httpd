package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

func Users(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	a := auth.NewAuthorizator(uid, role)
	objects := sys.users.AsObjects(a)

	return objects
}

func UpdateUsers(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	created, updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

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

	if err := save(TagUsers, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})

	return dbc.Objects(), nil
}

func GetUser(uid string) (auth.IUser, bool) {
	return sys.users.User(uid)
}

func SetPassword(uid, role, pwd string) error {
	sys.Lock()
	defer sys.Unlock()

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

	if updated, err := shadow.SetPassword(auth, uid, pwd, dbc); err != nil {
		return err
	} else {
		dbc.Stash(updated)
	}

	if err := shadow.Validate(); err != nil {
		return err
	}

	if err := save(TagUsers, &shadow); err != nil {
		return err
	}

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})

	return nil
}

func SetOTP(uid, role, secret string) error {
	sys.Lock()
	defer sys.Unlock()

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

	if updated, err := shadow.SetOTP(auth, uid, secret, dbc); err != nil {
		return err
	} else {
		dbc.Stash(updated)
	}

	if err := shadow.Validate(); err != nil {
		return err
	}

	if err := save(TagUsers, &shadow); err != nil {
		return err
	}

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})

	return nil
}

func RevokeOTP(uid, role string) error {
	sys.Lock()
	defer sys.Unlock()

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

	if updated, err := shadow.RevokeOTP(auth, uid, dbc); err != nil {
		return err
	} else {
		dbc.Stash(updated)
	}

	if err := shadow.Validate(); err != nil {
		return err
	}

	if err := save(TagUsers, &shadow); err != nil {
		return err
	}

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})

	return nil
}

func HasAdmin(role string) bool {
	return sys.users.HasAdmin(role)
}

func MakeAdminUser(name, uid, pwd, role string) error {
	sys.Lock()
	defer sys.Unlock()

	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

	if err := shadow.MakeAdminUser(name, uid, pwd, role, dbc); err != nil {
		return err
	}

	if err := save(TagUsers, &shadow); err != nil {
		return err
	}

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})

	return nil
}

func UserLogin(uid, role string, err error) {
	sys.Lock()
	defer sys.Unlock()

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.users.Clone()

	shadow.UserLogin(auth, uid, err, dbc)

	dbc.Commit(&sys, func() {
		sys.users = shadow
	})
}
