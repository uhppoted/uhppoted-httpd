package system

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type object catalog.Object

func UpdateControllers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	objects, err := unpack(m)
	if err != nil {
		return nil, err
	}

	list := struct {
		Objects []interface{} `json:"objects,omitempty"`
	}{}

	shadow := sys.controllers.Clone()

	for _, object := range objects {
		if updated, err := shadow.UpdateByOID(auth, object.OID, object.Value); err != nil {
			return nil, err
		} else if updated != nil {
			list.Objects = append(list.Objects, updated...)
		}
	}

	if err := save(shadow); err != nil {
		return nil, err
	}

	sys.controllers = *shadow

	sys.taskQ.Add(Task{
		f: func() {
			if err := controllers.Export(sys.conf, shadow.Controllers, sys.doors.Doors); err != nil {
				warn(err)
			}
		},
	})

	sys.taskQ.Add(Task{
		f: func() {
			info("Updating controllers from configuration")
			sys.controllers.Sync()
			UpdateACL()
		},
	})

	return list, nil
}

func save(c *controllers.ControllerSet) error {
	if err := validate(c); err != nil {
		return err
	}

	return c.Save()
}

func validate(c *controllers.ControllerSet) error {
	if err := c.Validate(); err != nil {
		return types.BadRequest(err, err)
	}

	doors := map[string]string{}

	for _, r := range c.Controllers {
		for _, v := range r.Doors {
			if v != "" {
				if _, ok := sys.doors.Doors[v]; !ok {
					return types.BadRequest(fmt.Errorf("Invalid door ID"), fmt.Errorf("controller %v: invalid door ID (%v)", r.OID, v))
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				return types.BadRequest(fmt.Errorf("%v door assigned to more than one controller", sys.doors.Doors[v].Name), fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID))
			}

			doors[v] = r.OID
		}
	}

	return nil
}

func unpack(m map[string]interface{}) ([]object, error) {
	f := func(err error) error {
		return types.BadRequest(fmt.Errorf("Invalid request (%v)", err), fmt.Errorf("Error unpacking 'post' request (%w)", err))
	}

	o := struct {
		Objects []object `json:"objects"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, f(err)
	}

	log.Printf("INFO %v", fmt.Sprintf("UNPACK %s\n", string(blob)))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, f(err)
	}

	return o.Objects, nil
}
