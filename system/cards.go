package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func Cards() interface{} {
	sys.RLock()
	defer sys.RUnlock()

	objects := sys.cards.AsObjects()

	return objects
}

func UpdateCards(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.cards.Cards.Clone()

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

	if err := save(sys.cards.file, sys.cards.tag, &shadow); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.cards.Cards = shadow
	sys.updated()

	list := squoosh(dbc.Objects())
	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: list,
	}, nil
}
