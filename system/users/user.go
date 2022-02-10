package users

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type User struct {
	OID  catalog.OID
	Name string
	UID  string
	Role string

	created core.DateTime
	deleted core.DateTime
}

type kv = struct {
	field catalog.Suffix
	value interface{}
}

const BLANK = "'blank'"

var created = core.DateTimeNow()

func (u User) IsValid() bool {
	if strings.TrimSpace(u.Name) != "" || strings.TrimSpace(u.UID) != "" {
		return true
	}

	return false
}

func (u User) IsDeleted() bool {
	return !u.deleted.IsZero()
}

func (u User) String() string {
	name := strings.TrimSpace(u.Name)
	if name != "" {
		return name
	}

	uid := strings.TrimSpace(u.UID)
	if uid != "" {
		return uid
	}

	return ""
}

func (u User) AsObjects(auth auth.OpAuth) []catalog.Object {
	list := []kv{}

	if u.IsDeleted() {
		list = append(list, kv{UserDeleted, u.deleted})
	} else {
		list = append(list, kv{UserStatus, u.status()})
		list = append(list, kv{UserCreated, u.created})
		list = append(list, kv{UserDeleted, u.deleted})
		list = append(list, kv{UserName, u.Name})
		list = append(list, kv{UserUID, u.UID})
		list = append(list, kv{UserRole, u.Role})
		list = append(list, kv{UserPassword, ""})
	}

	return u.toObjects(list, auth)
}

func (u User) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Name string
		UID  string
		Role string
	}{
		Name: u.Name,
		UID:  u.UID,
		Role: u.Role,
	}

	return "user", &entity
}

func (u *User) set(a auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if u == nil {
		return []catalog.Object{}, nil
	}

	if u.IsDeleted() {
		return u.toObjects([]kv{{UserDeleted, u.deleted}}, a), fmt.Errorf("User has been deleted")
	}

	f := func(field string, value interface{}) error {
		if a != nil {
			return a.CanUpdate(u, field, value, auth.Users)
		}

		return nil
	}

	list := []kv{}
	clone := u.clone()

	switch {
	case oid == u.OID.Append(UserName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			u.Name = strings.TrimSpace(value)
			list = append(list, kv{UserName, stringify(u.Name, "")})

			u.log(a,
				"update",
				u.OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(clone.Name, BLANK), stringify(u.Name, BLANK)),
				stringify(u.Name, ""),
				stringify(value, ""),
				dbc)
		}

	case oid == u.OID.Append(UserUID):
		if err := f("uid", value); err != nil {
			return nil, err
		} else {
			u.UID = strings.TrimSpace(value)
			list = append(list, kv{UserUID, stringify(u.UID, "")})

			u.log(a,
				"update",
				u.OID,
				"uid",
				fmt.Sprintf("Updated UID from %v to %v", stringify(clone.UID, BLANK), stringify(u.UID, BLANK)),
				stringify(clone.UID, ""),
				stringify(u.UID, ""),
				dbc)
		}

	case oid == u.OID.Append(UserRole):
		if err := f("role", value); err != nil {
			return nil, err
		} else {
			u.Role = strings.TrimSpace(value)
			list = append(list, kv{UserRole, stringify(u.Role, "")})

			u.log(a,
				"update",
				u.OID,
				"role",
				fmt.Sprintf("Updated role from %v to %v", stringify(clone.Role, BLANK), stringify(u.Role, BLANK)),
				stringify(clone.Role, ""),
				stringify(u.Role, ""),
				dbc)
		}
	}

	if strings.TrimSpace(u.Name) == "" && strings.TrimSpace(u.UID) == "" {
		if a != nil {
			if err := a.CanDelete(clone, auth.Users); err != nil {
				return nil, err
			}
		}

		if p := stringify(clone.UID, ""); p != "" {
			u.log(a, "delete", u.OID, "user", fmt.Sprintf("Deleted UID %v", p), "", "", dbc)
		} else if p = stringify(clone.Name, ""); p != "" {
			u.log(a, "delete", u.OID, "user", fmt.Sprintf("Deleted user %v", p), "", "", dbc)
		} else {
			u.log(a, "delete", u.OID, "user", "Deleted user", "", "", dbc)
		}

		u.deleted = core.DateTimeNow()
		list = append(list, kv{UserDeleted, u.deleted})

		catalog.Delete(u.OID)
	}

	list = append(list, kv{UserStatus, u.status()})

	return u.toObjects(list, a), nil
}

func (u User) toObjects(list []kv, a auth.OpAuth) []catalog.Object {
	f := func(u User, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(u, field, value, auth.Cards); err != nil {
				return false
			}
		}

		return true
	}

	objects := []catalog.Object{}

	if !u.IsDeleted() && f(u, "OID", u.OID) {
		objects = append(objects, catalog.NewObject(u.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(u, field, v.value) {
			objects = append(objects, catalog.NewObject2(u.OID, v.field, v.value))
		}
	}

	return objects
}

func (u User) status() types.Status {
	if u.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (u User) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		UID     string        `json:"uid,omitempty"`
		Role    string        `json:"role,omitempty"`
		Created core.DateTime `json:"created"`
	}{
		OID:     u.OID,
		Name:    strings.TrimSpace(u.Name),
		UID:     strings.TrimSpace(u.UID),
		Role:    strings.TrimSpace(u.Role),
		Created: u.created,
	}

	return json.Marshal(record)
}

func (u *User) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		UID     string        `json:"uid,omitempty"`
		Role    string        `json:"role,omitempty"`
		Created core.DateTime `json:"created"`
	}{
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	u.OID = record.OID
	u.Name = strings.TrimSpace(record.Name)
	u.UID = strings.TrimSpace(record.UID)
	u.Role = strings.TrimSpace(record.Role)
	u.created = record.Created

	return nil
}

func (u User) clone() *User {
	return &User{
		OID:  u.OID,
		Name: u.Name,
		UID:  u.UID,
		Role: u.Role,

		created: u.created,
		deleted: u.deleted,
	}
}

func (u User) log(auth auth.OpAuth, operation string, oid catalog.OID, field, description, before, after string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       oid,
		Component: "user",
		Operation: operation,
		Details: audit.Details{
			ID:          stringify(u.UID, ""),
			Name:        stringify(u.Name, ""),
			Field:       field,
			Description: description,
			Before:      before,
			After:       after,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}

func stringify(i interface{}, defval string) string {
	s := ""

	switch v := i.(type) {
	default:
		if v != nil {
			s = fmt.Sprintf("%v", v)
		}
	}

	if s != "" {
		return s
	}

	return defval
}
