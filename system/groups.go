package system

import (
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

	list := []catalog.Object{}
	dbc := db.NewDBC(sys.trail)
	shadow := sys.groups.Clone()

	for _, object := range objects {
		if updated, err := shadow.UpdateByOID(auth, object.OID, object.Value.(string), dbc); err != nil {
			return nil, err
		} else if updated != nil {
			list = append(list, updated...)
		}
	}

	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	if err := shadow.Save(); err != nil {
		return nil, err
	}

	dbc.Commit(list)
	sys.groups = shadow
	sys.updated()

	return struct {
		Objects []catalog.Object `json:"objects,omitempty"`
	}{
		Objects: list,
	}, nil
}
