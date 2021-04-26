package system

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/controllers"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type object struct {
	OID   string `json:"OID"`
	Value string `json:"value"`
}

func UpdateControllers(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	sys.Lock()

	defer sys.Unlock()

	// add/update ?

	objects, clist, err := unpack(m)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid request (%v)", err),
			Detail: fmt.Errorf("Error unpacking 'post' request (%w)", err),
		}
	}

	list := struct {
		Objects []interface{} `json:"objects,omitempty"`
		Added   []interface{} `json:"added,omitempty"`
		Updated []interface{} `json:"updated,omitempty"`
		Deleted []interface{} `json:"deleted,omitempty"`
	}{}

	shadow := sys.controllers.Clone()

	// Update objects
	for _, object := range objects {
		if view, err := shadow.UpdateByOID(object.OID, object.Value); err != nil {
			return nil, err
		} else if view != nil {
			list.Objects = append(list.Objects, view)
		}
	}

	// Add/update/delete controllers
loop:
	for _, c := range clist {
		// ... delete?
		if c.OID != "" && (c.Name == nil || *c.Name == "") && (c.DeviceID == nil || *c.DeviceID == 0) {
			for _, v := range shadow.Controllers {
				if v.OID == c.OID {
					if r, err := sys.delete(shadow, c, auth); err != nil {
						return nil, err
					} else if r != nil {
						if view := r.AsView(); view != nil {
							list.Deleted = append(list.Deleted, view)
						}
					}
				}
			}

			continue loop
		}

		// ... update controller?
		for _, v := range shadow.Controllers {
			if v.OID == c.OID {
				if r, err := sys.update(shadow, c, auth); err != nil {
					return nil, err
				} else if r != nil {
					if view := r.AsView(); view != nil {
						list.Updated = append(list.Updated, view)
					}
				}

				continue loop
			}
		}

		// ... add controller
		if r, err := sys.add(shadow, c, auth); err != nil {
			return nil, err
		} else if r != nil {
			if view := r.AsView(); view != nil {
				list.Added = append(list.Added, view)
			}
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

func unpack(m map[string]interface{}) ([]object, []controllers.Controller, error) {
	o := struct {
		Objects []object `json:"objects"`

		ControllerSet []struct {
			ID       string
			OID      *string
			Name     *string
			DeviceID *uint32
			IP       *string
			Doors    map[uint8]string
			DateTime *string
		} `json:"controllers"`
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("INFO %v", fmt.Sprintf("UNPACK %s\n", string(blob)))

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, nil, err
	}

	list := []controllers.Controller{}

	for _, r := range o.ControllerSet {
		record := controllers.Controller{}

		if r.OID != nil {
			record.OID = *r.OID
		}

		if r.Name != nil {
			name := types.Name(*r.Name)
			record.Name = &name
		}

		if r.DeviceID != nil {
			record.DeviceID = r.DeviceID
		}

		if r.IP != nil && *r.IP != "" {
			if addr, err := types.ResolveAddr(*r.IP); err != nil {
				return nil, nil, err
			} else {
				record.IP = addr
			}
		}

		if r.DateTime != nil {
			if tz, err := types.Timezone(strings.TrimSpace(*r.DateTime)); err != nil {
				return nil, nil, err
			} else {
				tzs := tz.String()
				record.TimeZone = &tzs
			}
		}

		if r.Doors != nil && len(r.Doors) > 0 {
			record.Doors = map[uint8]string{}
			for k, v := range r.Doors {
				record.Doors[k] = v
			}
		}

		list = append(list, record)
	}

	return o.Objects, list, nil
}
