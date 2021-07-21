package system

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid request (%v)", err),
			Detail: fmt.Errorf("Error unpacking 'post' request (%w)", err),
		}
	}

	list := struct {
		Objects []interface{} `json:"objects,omitempty"`
	}{}

	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	shadow := sys.controllers.Clone()

	for _, object := range objects {
		if updated, err := shadow.UpdateByOID(uid, object.OID, object.Value); err != nil {
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

func (s *system) add(shadow *controllers.ControllerSet, c controllers.Controller, auth auth.OpAuth) (*controllers.Controller, error) {
	if auth != nil {
		if err := auth.CanAddController(&c); err != nil {
			return nil, &types.HttpdError{
				Status: http.StatusUnauthorized,
				Err:    fmt.Errorf("Not authorized to add controller"),
				Detail: err,
			}
		}
	}

	record, err := shadow.Add(c)
	if err != nil {
		return nil, err
	}

	s.log("add", record, auth)

	return record, nil
}

func (s *system) update(shadow *controllers.ControllerSet, c controllers.Controller, auth auth.OpAuth) (*controllers.Controller, error) {
	var current *controllers.Controller

	for _, v := range s.controllers.Controllers {
		if v.OID == c.OID {
			current = v
			break
		}
	}

	record, err := shadow.Update(c)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    err,
			Detail: fmt.Errorf("Invalid 'update' request (%w)", err),
		}
	}

	if auth != nil {
		if err := auth.CanUpdateController(current, record); err != nil {
			return nil, &types.HttpdError{
				Status: http.StatusUnauthorized,
				Err:    fmt.Errorf("Not authorized to update controller"),
				Detail: err,
			}
		}
	}

	s.log("update", map[string]interface{}{"original": current, "updated": record}, auth)

	return record, nil
}

func (s *system) delete(shadow *controllers.ControllerSet, c controllers.Controller, auth auth.OpAuth) (*controllers.Controller, error) {
	record, err := shadow.Delete(c)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusUnauthorized,
			Err:    err,
			Detail: fmt.Errorf("Invalid 'update' request (%w)", err),
		}
	}

	if record != nil && auth != nil {
		if err := auth.CanDeleteController(record); err != nil {
			return nil, &types.HttpdError{
				Status: http.StatusUnauthorized,
				Err:    fmt.Errorf("Not authorized to delete controller"),
				Detail: fmt.Errorf("Invalid 'update' request (%w)", fmt.Errorf("Not authorized to delete controller")),
			}
		}
	}

	catalog.Delete(record.OID)

	s.log("delete", record, auth)

	return record, nil
}

func save(c *controllers.ControllerSet) error {
	if err := validate(c); err != nil {
		return err
	}

	return c.Save()
}

func validate(c *controllers.ControllerSet) error {
	if err := c.Validate(); err != nil {
		return err
	}

	doors := map[string]string{}

	for _, r := range c.Controllers {
		for _, v := range r.Doors {
			if v != "" {
				if _, ok := sys.doors.Doors[v]; !ok {
					return &types.HttpdError{
						Status: http.StatusBadRequest,
						Err:    fmt.Errorf("Invalid door ID"),
						Detail: fmt.Errorf("controller %v: invalid door ID (%v)", r.OID, v),
					}
				}
			}

			if rid, ok := doors[v]; ok && v != "" {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("%v door assigned to more than one controller", sys.doors.Doors[v].Name),
					Detail: fmt.Errorf("door %v: assigned to controllers %v and %v", v, rid, r.OID),
				}
			}

			doors[v] = r.OID
		}
	}

	return nil
}

func unpack(m map[string]interface{}) ([]object, error) {
	o := struct {
		Objects []object `json:"objects"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	log.Printf("INFO %v", fmt.Sprintf("UNPACK %s\n", string(blob)))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, err
	}

	return o.Objects, nil
}
