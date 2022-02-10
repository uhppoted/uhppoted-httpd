package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Users struct {
	users map[catalog.OID]*User
}

var guard sync.RWMutex

func NewUsers() Users {
	return Users{
		users: map[catalog.OID]*User{},
	}
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
				return fmt.Errorf("user '%v': duplicate OID (%v)", u.Name, u.OID)
			}

			uu.users[u.OID] = &u
		}
	}

	for _, u := range uu.users {
		catalog.PutUser(u.OID)
		catalog.PutV(u.OID, catalog.UserName, u.Name)
		catalog.PutV(u.OID, catalog.UserUID, u.UID)
		catalog.PutV(u.OID, catalog.UserRole, u.Role)
	}

	return nil
}

func (uu Users) Save() (json.RawMessage, error) {
	if err := validate(uu); err != nil {
		return nil, err
	}

	if err := uu.scrub(); err != nil {
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
		users: map[catalog.OID]*User{},
	}

	for oid, u := range uu.users {
		shadow.users[oid] = u.clone()
	}

	return shadow
}

func (uu *Users) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if uu == nil {
		return nil, nil
	}

	for k, u := range uu.users {
		if u.OID.Contains(oid) {
			objects, err := u.set(auth, oid, value, dbc)
			if err == nil {
				uu.users[k] = u
			}

			return objects, err
		}
	}

	objects := []catalog.Object{}

	if oid == "<new>" {
		if u, err := uu.add(auth, User{}); err != nil {
			return nil, err
		} else if u == nil {
			return nil, fmt.Errorf("Failed to add 'new' user")
		} else {
			uu.users[u.OID] = u
			objects = append(objects, catalog.NewObject(u.OID, "new"))
			objects = append(objects, catalog.NewObject2(u.OID, UserCreated, u.created))

			u.log(auth, "add", u.OID, "card", "Added 'new' user", "", "", dbc)
		}
	}

	return objects, nil
}

func (uu Users) Validate() error {
	return validate(uu)
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

func (uu *Users) AsObjects(auth auth.OpAuth) []catalog.Object {
	objects := []catalog.Object{}
	guard.RLock()

	defer guard.RUnlock()

	for _, u := range uu.users {
		if u.IsValid() || u.IsDeleted() {
			if l := u.AsObjects(auth); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (uu Users) Sweep(retention time.Duration) {
	cutoff := time.Now().Add(-retention)
	for k, u := range uu.users {
		if u.IsDeleted() && u.deleted.Before(cutoff) {
			delete(uu.users, k)
		}
	}
}

func (uu Users) add(a auth.OpAuth, u User) (*User, error) {
	oid := catalog.NewUser()
	if _, ok := uu.users[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	user := u.clone()
	user.OID = oid
	user.created = core.DateTimeNow()

	if a != nil {
		if err := a.CanAdd(user, auth.Users); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func validate(uu Users) error {
	users := map[string]catalog.OID{}

	for _, u := range uu.users {
		if u.IsDeleted() {
			continue
		}

		if u.OID == "" {
			return fmt.Errorf("Invalid user OID (%v)", u.OID)
		}

		if oid, ok := users[u.UID]; ok {
			return &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Duplicate UID (%v)", u.UID),
				Detail: fmt.Errorf("UID %v: duplicate entry in records %v and %v", u.UID, oid, u.OID),
			}
		}

		users[u.UID] = u.OID
	}

	return nil
}

func (uu *Users) scrub() error {
	return nil
}
