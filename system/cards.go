package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Cards(uid, role string) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.cards.AsObjects(auth)

	return objects
}

func UpdateCards(uid, role string, m map[string]interface{}) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.cards.Cards.Clone()

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

	if err := save(sys.cards.file, sys.cards.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.cards.Cards = shadow
		sys.updated()
	})

	return dbc.Objects(), nil
}
