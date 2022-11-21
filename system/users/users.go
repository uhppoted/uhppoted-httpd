package users

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Users struct {
	users map[schema.OID]*User
}

var guard sync.RWMutex

func NewUsers() Users {
	return Users{
		users: map[schema.OID]*User{},
	}
}

func (uu *Users) AsObjects(a *auth.Authorizator) []schema.Object {
	objects := []schema.Object{}
	guard.RLock()

	defer guard.RUnlock()

	for _, u := range uu.users {
		if u.IsValid() || u.IsDeleted() {
			catalog.Join(&objects, u.AsObjects(a)...)
		}
	}

	return objects
}

func (uu *Users) Create(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if uu != nil {
		if u, err := uu.add(a, User{}); err != nil {
			return nil, err
		} else if u == nil {
			return nil, fmt.Errorf("Failed to add 'new' user")
		} else {
			u.log(dbc, auth.UID(a), "add", "user", "", "", "Added 'new' user")

			catalog.Join(&objects, catalog.NewObject(u.OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(u.OID, UserCreated, u.created))
		}
	}

	return objects, nil
}

func (uu *Users) Update(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if uu != nil {
		for k, u := range uu.users {
			if u.OID.Contains(oid) {
				objects, err := u.set(a, oid, value, dbc)
				if err == nil {
					uu.users[k] = u
				}

				return objects, err
			}
		}
	}

	return objects, nil
}

func (uu *Users) Delete(a *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if uu != nil {
		for k, u := range uu.users {
			if u.OID == oid {
				objects, err := u.delete(a, dbc)
				if err == nil {
					uu.users[k] = u
				}

				return objects, err
			}
		}
	}

	return objects, nil
}

func (uu *Users) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var u User
		if err := u.deserialize(v); err == nil {
			if _, ok := uu.users[u.OID]; ok {
				return fmt.Errorf("user '%v': duplicate OID (%v)", u.name, u.OID)
			}

			uu.users[u.OID] = &u
		}
	}

	for _, u := range uu.users {
		catalog.PutT(u.CatalogUser)
		catalog.PutV(u.OID, schema.UserName, u.name)
		catalog.PutV(u.OID, schema.UserUID, u.uid)
		catalog.PutV(u.OID, schema.UserRole, u.role)
	}

	return nil
}

func (uu Users) Save() (json.RawMessage, error) {
	if err := uu.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, u := range uu.users {
		if u.IsValid() && !u.IsDeleted() {
			if record, err := u.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (uu Users) Clone() Users {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Users{
		users: map[schema.OID]*User{},
	}

	for oid, u := range uu.users {
		shadow.users[oid] = u.clone()
	}

	return shadow
}

func (uu *Users) SetPassword(a *auth.Authorizator, uid, pwd string, dbc db.DBC) ([]schema.Object, error) {
	if uu == nil {
		return nil, nil
	}

	for k, u := range uu.users {
		if u.uid == uid {
			objects, err := u.set(a, u.OID.Append(UserPassword), pwd, dbc)
			if err == nil {
				uu.users[k] = u
			}

			return objects, err
		}
	}

	return []schema.Object{}, nil
}

func (uu *Users) GetOTP(a *auth.Authorizator, uid string) (string, error) {
	if uu == nil {
		return "", nil
	}

	for _, u := range uu.users {
		if u.uid == uid {
			return u.get(a, u.OID.Append(UserOTPKey))
		}
	}

	return "", fmt.Errorf("invalid UID (%v)", uid)
}

func (uu *Users) SetOTP(a *auth.Authorizator, uid, secret string, dbc db.DBC) ([]schema.Object, error) {
	if uu == nil {
		return nil, nil
	}

	for k, u := range uu.users {
		if u.uid == uid {
			objects, err := u.set(a, u.OID.Append(UserOTPKey), secret, dbc)
			if err == nil {
				uu.users[k] = u
			}

			return objects, err
		}
	}

	return []schema.Object{}, nil
}

func (uu *Users) RevokeOTP(a *auth.Authorizator, uid string, dbc db.DBC) ([]schema.Object, error) {
	if uu == nil {
		return nil, nil
	}

	for k, u := range uu.users {
		if u.uid == uid {
			objects, err := u.set(a, u.OID.Append(UserOTPKey), "", dbc)
			if err == nil {
				uu.users[k] = u
			}

			return objects, err
		}
	}

	return []schema.Object{}, nil
}

func (uu Users) User(uid string) (auth.IUser, bool) {
	if strings.TrimSpace(uid) != "" {
		for _, u := range uu.users {
			if u.uid == uid {
				return u, true
			}
		}
	}

	return nil, false
}

func (uu Users) Validate() error {
	users := map[string]schema.OID{}

	for k, u := range uu.users {
		if u.IsDeleted() {
			continue
		}

		if u.OID == "" {
			return fmt.Errorf("Invalid user OID (%v)", u.OID)
		} else if u.OID != k {
			return fmt.Errorf("User %s: mismatched user OID %v (expected %v)", u.name, u.OID, k)
		}

		if err := u.validate(); err != nil {
			if !u.modified.IsZero() {
				return err
			}
		}

		if _, ok := users[u.uid]; ok {
			return fmt.Errorf("Duplicate UID (%v)", u.uid)
		}

		if u.uid != "" {
			users[u.uid] = u.OID
		}
	}

	return nil
}

func (uu Users) Print() {
	serializable := []json.RawMessage{}
	for _, u := range uu.users {
		if u.IsValid() && !u.IsDeleted() {
			if record, err := u.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- USERS\n%s\n", string(b))
	}
}

func (uu *Users) Sweep(retention time.Duration) {
	cutoff := time.Now().Add(-retention)
	for k, u := range uu.users {
		if u.IsDeleted() && u.deleted.Before(cutoff) {
			delete(uu.users, k)
		}
	}
}

func (uu *Users) add(a auth.OpAuth, u User) (*User, error) {
	oid := catalog.NewT(u.CatalogUser)
	if _, ok := uu.users[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	user := u.clone()
	user.OID = oid
	user.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(user, auth.Users); err != nil {
			return nil, err
		}
	}

	uu.users[user.OID] = user

	return user, nil
}
