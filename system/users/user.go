package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type User struct {
	ctypes.CatalogUser
	OID      schema.OID
	name     string
	uid      string
	role     string
	salt     []byte
	password string

	created  types.Timestamp
	deleted  types.Timestamp
	modified types.Timestamp
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

const BLANK = "'blank'"

var created = types.TimestampNow()

func (u User) IsValid() bool {
	if strings.TrimSpace(u.name) != "" || strings.TrimSpace(u.uid) != "" {
		return true
	}

	return false
}

func (u User) IsDeleted() bool {
	return !u.deleted.IsZero()
}

func (u User) Password() ([]byte, string) {
	salt := make([]byte, len(u.salt))
	copy(salt, u.salt)

	return salt, u.password
}

func (u User) Role() string {
	return u.role
}

func (u User) String() string {
	name := strings.TrimSpace(u.name)
	if name != "" {
		return name
	}

	uid := strings.TrimSpace(u.uid)
	if uid != "" {
		return uid
	}

	return ""
}

func (u User) AsObjects(auth auth.OpAuth) []schema.Object {
	list := []kv{}

	if u.IsDeleted() {
		list = append(list, kv{UserDeleted, u.deleted})
	} else {
		list = append(list, kv{UserStatus, u.status()})
		list = append(list, kv{UserCreated, u.created})
		list = append(list, kv{UserDeleted, u.deleted})
		list = append(list, kv{UserName, u.name})
		list = append(list, kv{UserUID, u.uid})
		list = append(list, kv{UserRole, u.role})
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
		Name: u.name,
		UID:  u.uid,
		Role: u.role,
	}

	return "user", &entity
}

func (u *User) set(a auth.OpAuth, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if u == nil {
		return []schema.Object{}, nil
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
			u.name = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserName, stringify(u.name, "")})

			u.log(a,
				"update",
				u.OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(clone.name, BLANK), stringify(u.name, BLANK)),
				stringify(u.name, ""),
				stringify(value, ""),
				dbc)
		}

	case oid == u.OID.Append(UserUID):
		if err := f("uid", value); err != nil {
			return nil, err
		} else {
			u.uid = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserUID, stringify(u.uid, "")})

			u.log(a,
				"update",
				u.OID,
				"uid",
				fmt.Sprintf("Updated UID from %v to %v", stringify(clone.uid, BLANK), stringify(u.uid, BLANK)),
				stringify(clone.uid, ""),
				stringify(u.uid, ""),
				dbc)
		}

	case oid == u.OID.Append(UserRole):
		if err := f("role", value); err != nil {
			return nil, err
		} else {
			u.role = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserRole, stringify(u.role, "")})

			u.log(a,
				"update",
				u.OID,
				"role",
				fmt.Sprintf("Updated role from %v to %v", stringify(clone.role, BLANK), stringify(u.role, BLANK)),
				stringify(clone.role, ""),
				stringify(u.role, ""),
				dbc)
		}

	case oid == u.OID.Append(UserPassword):
		if err := f("password", value); err != nil {
			return nil, err
		} else {
			salt := make([]byte, 16)
			if _, err := io.ReadFull(rand.Reader, salt); err != nil {
				return nil, err
			}

			h := sha256.New()
			h.Write(salt)
			h.Write([]byte(value))

			u.salt = salt
			u.password = fmt.Sprintf("%0x", h.Sum(nil))
			u.modified = types.TimestampNow()

			list = append(list, kv{UserPassword, ""})

			u.log(a, "update", u.OID, "password", "Updated password", "", "", dbc)
		}
	}

	if strings.TrimSpace(u.name) == "" && strings.TrimSpace(u.uid) == "" {
		if a != nil {
			if err := a.CanDelete(clone, auth.Users); err != nil {
				return nil, err
			}
		}

		if p := stringify(clone.uid, ""); p != "" {
			u.log(a, "delete", u.OID, "user", fmt.Sprintf("Deleted UID %v", p), "", "", dbc)
		} else if p = stringify(clone.name, ""); p != "" {
			u.log(a, "delete", u.OID, "user", fmt.Sprintf("Deleted user %v", p), "", "", dbc)
		} else {
			u.log(a, "delete", u.OID, "user", "Deleted user", "", "", dbc)
		}

		u.deleted = types.TimestampNow()
		u.modified = types.TimestampNow()
		list = append(list, kv{UserDeleted, u.deleted})

		catalog.DeleteT(u, u.OID)
	}

	list = append(list, kv{UserStatus, u.status()})

	return u.toObjects(list, a), nil
}

func (u User) toObjects(list []kv, a auth.OpAuth) []schema.Object {
	f := func(u User, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(u, field, value, auth.Cards); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if !u.IsDeleted() && f(u, "OID", u.OID) {
		catalog.Join(&objects, catalog.NewObject(u.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(u, field, v.value) {
			catalog.Join(&objects, catalog.NewObject2(u.OID, v.field, v.value))
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
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		UID      string          `json:"uid,omitempty"`
		Role     string          `json:"role,omitempty"`
		Salt     string          `json:"salt"`
		Password string          `json:"password"`
		Created  types.Timestamp `json:"created"`
		Modified types.Timestamp `json:"modified"`
	}{
		OID:      u.OID,
		Name:     strings.TrimSpace(u.name),
		UID:      strings.TrimSpace(u.uid),
		Role:     strings.TrimSpace(u.role),
		Salt:     hex.EncodeToString(u.salt[:]),
		Password: u.password,
		Created:  u.created,
		Modified: u.modified,
	}

	return json.Marshal(record)
}

func (u *User) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		UID      string          `json:"uid,omitempty"`
		Role     string          `json:"role,omitempty"`
		Salt     string          `json:"salt"`
		Password string          `json:"password"`
		Created  types.Timestamp `json:"created"`
		Modified types.Timestamp `json:"modified"`
	}{
		Created:  created,
		Modified: types.TimestampNow(),
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	salt, err := hex.DecodeString(record.Salt)
	if err != nil {
		return err
	}

	u.OID = record.OID
	u.name = strings.TrimSpace(record.Name)
	u.uid = strings.TrimSpace(record.UID)
	u.role = strings.TrimSpace(record.Role)
	u.salt = salt
	u.password = record.Password
	u.created = record.Created
	u.modified = record.Modified

	return nil
}

func (u User) clone() *User {
	replicant := User{
		OID:      u.OID,
		name:     u.name,
		uid:      u.uid,
		role:     u.role,
		salt:     make([]byte, len(u.salt)),
		password: u.password,

		created: u.created,
		deleted: u.deleted,
	}

	copy(replicant.salt, u.salt)

	return &replicant
}

func (u User) log(auth auth.OpAuth, operation string, oid schema.OID, field, description, before, after string, dbc db.DBC) {
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
			ID:          stringify(u.uid, ""),
			Name:        stringify(u.name, ""),
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
