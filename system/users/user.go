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

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

const MaxFailed uint32 = 5

type User struct {
	catalog.CatalogUser
	name     string
	uid      string
	role     string
	salt     []byte
	password string
	otp      string
	locked   bool
	failed   uint32

	created  types.Timestamp
	deleted  types.Timestamp
	modified types.Timestamp
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

var created = types.TimestampNow()

func (u User) IsValid() bool {
	return u.validate() == nil
}

func (u User) validate() error {
	if strings.TrimSpace(u.name) == "" && strings.TrimSpace(u.uid) == "" {
		return fmt.Errorf("User name and user ID cannot both be blank")
	}

	return nil
}

func (u User) IsDeleted() bool {
	return !u.deleted.IsZero()
}

func (u User) Password() ([]byte, string) {
	salt := make([]byte, len(u.salt))
	copy(salt, u.salt)

	return salt, u.password
}

func (u User) OTPKey() string {
	return u.otp
}

func (u User) Role() string {
	return u.role
}

func (u User) Locked() bool {
	return u.locked
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

func (u User) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if u.IsDeleted() {
		list = append(list, kv{UserDeleted, u.deleted})
	} else {
		list = append(list, kv{UserStatus, u.Status()})
		list = append(list, kv{UserCreated, u.created})
		list = append(list, kv{UserDeleted, u.deleted})
		list = append(list, kv{UserName, u.name})
		list = append(list, kv{UserUID, u.uid})
		list = append(list, kv{UserRole, u.role})
		list = append(list, kv{UserPassword, ""})
		list = append(list, kv{UserOTP, u.otp != ""})
		list = append(list, kv{UserLocked, u.locked})
	}

	return u.toObjects(list, a)
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

func (u User) get(a *auth.Authorizator, oid schema.OID) (string, error) {
	if u.IsDeleted() {
		return "", fmt.Errorf("User has been deleted")
	}

	switch {
	case oid == u.OID.Append(UserOTP):
		if err := CanView(a, u, "otp", ""); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%v", u.otp != ""), nil
		}
	}

	return "", nil
}

func (u *User) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
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

	uid := auth.UID(a)
	list := []kv{}
	clone := u.clone()

	switch {
	case oid == u.OID.Append(UserName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			u.name = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserName, u.name})

			u.log(dbc, uid, "update", "name", u.name, value, "Updated name from %v to %v", clone.name, u.name)
		}

	case oid == u.OID.Append(UserUID):
		if err := f("uid", value); err != nil {
			return nil, err
		} else {
			u.uid = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserUID, u.uid})

			u.log(dbc, uid, "update", "uid", clone.uid, u.uid, "Updated UID from %v to %v", clone.uid, u.uid)
		}

	case oid == u.OID.Append(UserRole):
		if err := f("role", value); err != nil {
			return nil, err
		} else {
			u.role = strings.TrimSpace(value)
			u.modified = types.TimestampNow()
			list = append(list, kv{UserRole, u.role})

			u.log(dbc, uid, "update", "role", clone.role, u.role, "Updated role from %v to %v", clone.role, u.role)
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

			u.log(dbc, uid, "update", "password", "", "", "Updated password")
		}

	// ... 'revoke only' from UI
	case oid == u.OID.Append(UserOTP):
		if err := f("otp", value); err != nil {
			return nil, err
		} else if value == "false" {
			u.otp = ""
			u.modified = types.TimestampNow()
			u.log(dbc, uid, "update", "otp", u.uid, "", "Revoked OTP")
		}

		list = append(list, kv{UserOTP, u.otp != ""})

	case oid == u.OID.Append(UserOTPKey):
		if err := f("otp", value); err != nil {
			return nil, err
		} else {
			if value == "" {
				u.log(dbc, uid, "update", "otp", "", "", "Revoked OTP for %v (%v)", u.uid, u.name)
			} else if u.otp == "" {
				u.log(dbc, uid, "update", "otp", "", "", "Enabled OTP")
			} else {
				u.log(dbc, uid, "update", "otp", "", "", "Updated OTP")
			}

			u.otp = value
			u.modified = types.TimestampNow()

			list = append(list, kv{UserOTP, u.otp != ""})
		}

	// ... 'unlock only' from UI
	case oid == u.OID.Append(UserLocked):
		if err := f("locked", value); err != nil {
			return nil, err
		} else if value == "false" {
			u.failed = 0
			u.locked = false
			u.modified = types.TimestampNow()
			u.log(dbc, uid, "update", "locked", "locked", "unlocked", "Unlocked account for %v (%v)", u.uid, u.name)
		}

		list = append(list, kv{UserLocked, u.locked})
	}

	list = append(list, kv{UserStatus, u.Status()})

	return u.toObjects(list, a), nil
}

func (u *User) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if u != nil {
		if a != nil {
			if err := a.CanDelete(u, auth.Users); err != nil {
				return nil, err
			}
		}

		uid := auth.UID(a)
		if u.uid != "" {
			u.log(dbc, uid, "delete", "user", u.uid, "", "Deleted UID %v", u.uid)
		} else if u.name != "" {
			u.log(dbc, uid, "delete", "user", u.name, "", "Deleted user %v", u.name)
		} else {
			u.log(dbc, uid, "delete", "user", "", "", "Deleted user")
		}

		u.deleted = types.TimestampNow()
		u.modified = types.TimestampNow()

		list = append(list, kv{UserDeleted, u.deleted})
		list = append(list, kv{UserStatus, u.Status()})

		catalog.DeleteT(u.CatalogUser, u.OID)
	}

	return u.toObjects(list, a), nil
}

func (u User) toObjects(list []kv, a auth.OpAuth) []schema.Object {
	objects := []schema.Object{}

	if !u.IsDeleted() {
		if err := CanView(a, u, "OID", u.OID); err == nil {
			catalog.Join(&objects, catalog.NewObject(u.OID, ""))
		}
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if err := CanView(a, u, field, v.value); err == nil {
			catalog.Join(&objects, catalog.NewObject2(u.OID, v.field, v.value))
		}
	}

	return objects
}

func (u User) Status() types.Status {
	if u.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (u *User) login(err error, dbc db.DBC) {
	if err != nil {
		u.failed++

		if u.failed >= MaxFailed && !u.locked {
			u.locked = true
			u.log(dbc, u.uid, "update", "user", u.name, "locked", "Too many failed logins")
		}
	} else {
		u.failed = 0
	}
}

func (u User) serialize() ([]byte, error) {
	record := struct {
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		UID      string          `json:"uid,omitempty"`
		Role     string          `json:"role,omitempty"`
		Salt     string          `json:"salt"`
		Password string          `json:"password"`
		OTP      string          `json:"otp,omitempty"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
	}{
		OID:      u.OID,
		Name:     strings.TrimSpace(u.name),
		UID:      strings.TrimSpace(u.uid),
		Role:     strings.TrimSpace(u.role),
		Salt:     hex.EncodeToString(u.salt[:]),
		Password: u.password,
		OTP:      u.otp,
		Created:  u.created.UTC(),
		Modified: u.modified.UTC(),
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
		OTP      string          `json:"otp,omitempty"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
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
	u.otp = record.OTP
	u.created = record.Created
	u.modified = record.Modified

	return nil
}

func (u User) clone() *User {
	replicant := User{
		CatalogUser: catalog.CatalogUser{
			OID: u.OID,
		},
		name:     u.name,
		uid:      u.uid,
		role:     u.role,
		salt:     make([]byte, len(u.salt)),
		password: u.password,
		otp:      u.otp,
		locked:   u.locked,
		failed:   u.failed,

		created: u.created,
		deleted: u.deleted,
	}

	copy(replicant.salt, u.salt)

	return &replicant
}

func (u User) log(dbc db.DBC, uid, op string, field string, before, after any, format string, fields ...any) {
	dbc.Log(uid, op, u.OID, "user", u.uid, u.name, field, before, after, format, fields...)
}
