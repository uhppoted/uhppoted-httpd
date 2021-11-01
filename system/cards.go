package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateCards(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	updated := []catalog.Object{}
	dbc := db.NewDBC(sys.trail)
	shadow := sys.cards.Clone()

	for _, object := range objects {
		if l, err := shadow.UpdateByOID(auth, object.OID, object.Value.(string), dbc); err != nil {
			return nil, err
		} else if l != nil {
			updated = append(updated, l...)
		}
	}

	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	if err := shadow.Save(); err != nil {
		return nil, err
	}

	dbc.Commit(updated)
	sys.cards = shadow
	sys.cards.Stash()
	sys.updated()

	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: updated,
	}, nil

}
