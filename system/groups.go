package system

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateGroups(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()
	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	dbc := db.NewDBC(sys.trail)
	shadow := sys.groups.groups.Clone()

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

	if bytes, err := shadow.Save(); err != nil {
		return nil, err
	} else if bytes == nil {
		return nil, fmt.Errorf("invalid serialized 'groups' (%v)", err)
	} else if err := save(sys.groups.file, sys.groups.tag, bytes); err != nil {
		return nil, err
	}

	dbc.Commit()
	sys.groups.groups = shadow
	sys.updated()

	list := squoosh(dbc.Objects())
	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: list,
	}, nil
}
