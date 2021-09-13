package system

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func UpdateCards(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	list := struct {
		Objects []interface{} `json:"objects,omitempty"`
	}{}

	shadow := sys.cards.Clone()

	for _, object := range objects {
		if updated, err := shadow.UpdateByOID(auth, object.OID, object.Value); err != nil {
			return nil, err
		} else if updated != nil {
			list.Objects = append(list.Objects, updated...)
		}
	}

	// ... validate
	if err := shadow.Validate(); err != nil {
		return nil, types.BadRequest(err, err)
	}

	//	// ... save
	//	if err := shadow.Save(); err != nil {
	//		return nil, err
	//	}

	sys.cards = shadow
	//	sys.updated()

	return list, nil
}
