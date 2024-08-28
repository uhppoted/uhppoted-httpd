package system

import (
	lib "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

func SetDefaultCardStartDate(v string) {
	if date, err := lib.ParseDate(v); err == nil && !date.IsZero() {
		sys.acl.defaultStartDate = date
		infof("cards", "default card start date %v", date)
	}
}

func SetDefaultCardEndDate(v string) {
	if date, err := lib.ParseDate(v); err == nil && !date.IsZero() {
		sys.acl.defaultEndDate = date
		infof("cards", "default card end date   %v", date)
	}
}

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

	created, updated, deleted, err := unpack(m)
	if err != nil {
		return nil, err
	}

	auth := auth.NewAuthorizator(uid, role)
	dbc := db.NewDBC(sys.trail)
	shadow := sys.cards.Clone()

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

	if err := save(TagCards, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit(&sys, func() {
		sys.cards = shadow
	})

	return dbc.Objects(), nil
}
