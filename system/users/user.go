package users

import (
	"encoding/json"
	"fmt"
	//	"regexp"
	//	"strconv"
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

// const BLANK = "'blank'"

var created = core.DateTimeNow()

func (u User) IsValid() bool {
	if u.UID != "" {
		return true
	}

	return false
}

func (u User) IsDeleted() bool {
	return !u.deleted.IsZero()
}

// func (c Card) String() string {
//     name := strings.TrimSpace(c.Name)
//     if name == "" {
//         name = "-"
//     }

//     number := "-"

//     if c.Card != nil {
//         number = fmt.Sprintf("%v", c.Card)
//     }

//     return fmt.Sprintf("%v (%v)", number, name)
// }

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
	//     if c == nil {
	//         return []catalog.Object{}, nil
	//     }

	//     if c.IsDeleted() {
	//         return c.toObjects([]kv{{CardDeleted, c.deleted}}, a), fmt.Errorf("Card has been deleted")
	//     }

	//     f := func(field string, value interface{}) error {
	//         if a != nil {
	//             return a.CanUpdate(c, field, value, auth.Cards)
	//         }

	//         return nil
	//     }

	list := []kv{}
	//     clone := c.clone()

	//     switch {
	//     case oid == c.OID.Append(CardName):
	//         if err := f("name", value); err != nil {
	//             return nil, err
	//         } else {
	//             c.log(a,
	//                 "update",
	//                 c.OID,
	//                 "name",
	//                 fmt.Sprintf("Updated name from %v to %v", stringify(c.Name, BLANK), stringify(value, BLANK)),
	//                 stringify(c.Name, ""),
	//                 stringify(value, ""),
	//                 dbc)

	//             c.Name = strings.TrimSpace(value)
	//             list = append(list, kv{CardName, stringify(c.Name, "")})
	//         }

	//     case oid == c.OID.Append(CardNumber):
	//         if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
	//             if n, err := strconv.ParseUint(value, 10, 32); err != nil {
	//                 return nil, err
	//             } else if err := f("number", n); err != nil {
	//                 return nil, err
	//             } else {
	//                 c.log(a,
	//                     "update",
	//                     c.OID,
	//                     "card",
	//                     fmt.Sprintf("Updated card number from %v to %v", c.Card, value),
	//                     stringify(c.Card, ""),
	//                     stringify(value, ""),
	//                     dbc)

	//                 v := types.Card(n)
	//                 c.Card = &v
	//                 list = append(list, kv{CardNumber, c.Card})
	//             }
	//         } else if value == "" {
	//             if err := f("number", 0); err != nil {
	//                 return nil, err
	//             } else {
	//                 if p := stringify(c.Name, ""); p != "" {
	//                     c.log(a,
	//                         "update",
	//                         c.OID,
	//                         "number",
	//                         fmt.Sprintf("Cleared card number %v for %v", c.Card, p),
	//                         stringify(c.Card, ""),
	//                         stringify(p, ""),
	//                         dbc)
	//                 } else {
	//                     c.log(a,
	//                         "update",
	//                         c.OID,
	//                         "number",
	//                         fmt.Sprintf("Cleared card number %v", c.Card),
	//                         stringify(c.Card, ""),
	//                         "",
	//                         dbc)
	//                 }

	//                 c.Card = nil
	//                 list = append(list, kv{CardNumber, ""})
	//             }
	//         }

	//     case oid == c.OID.Append(CardFrom):
	//         if err := f("from", value); err != nil {
	//             return nil, err
	//         } else if from, err := core.DateFromString(value); err != nil {
	//             return nil, err
	//         } else if !from.IsValid() {
	//             return nil, fmt.Errorf("invalid 'from' date (%v)", value)
	//         } else {
	//             c.log(a,
	//                 "update",
	//                 c.OID,
	//                 "from",
	//                 fmt.Sprintf("Updated VALID FROM date from %v to %v", c.From, value),
	//                 stringify(c.From, ""),
	//                 stringify(value, ""),
	//                 dbc)

	//             c.From = from
	//             list = append(list, kv{CardFrom, c.From})
	//         }

	//     case oid == c.OID.Append(CardTo):
	//         if err := f("to", value); err != nil {
	//             return nil, err
	//         } else if to, err := core.DateFromString(value); err != nil {
	//             return nil, err
	//         } else if to.IsValid() {
	//             return nil, fmt.Errorf("invalid 'to' date (%v)", value)
	//         } else {
	//             c.log(a,
	//                 "update",
	//                 c.OID,
	//                 "to",
	//                 fmt.Sprintf("Updated VALID UNTIL date from %v to %v", c.From, value),
	//                 stringify(c.From, ""),
	//                 stringify(value, ""),
	//                 dbc)

	//             c.To = to
	//             list = append(list, kv{CardTo, c.To})
	//         }

	//     case catalog.OID(c.OID.Append(CardGroups)).Contains(oid):
	//         if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
	//             gid := m[1]
	//             k := catalog.GroupsOID.AppendS(gid)

	//             if err := f("group", value); err != nil {
	//                 return nil, err
	//             } else if !catalog.HasGroup(catalog.OID(k)) {
	//                 return nil, fmt.Errorf("invalid group OID (%v)", k)
	//             } else {
	//                 group := catalog.GetV(catalog.OID(k), GroupName)

	//                 if value == "true" {
	//                     c.log(a,
	//                         "update",
	//                         c.OID,
	//                         "group",
	//                         fmt.Sprintf("Granted access to %v", group),
	//                         "",
	//                         "",
	//                         dbc)
	//                 } else {
	//                     c.log(a,
	//                         "update",
	//                         c.OID,
	//                         "group",
	//                         fmt.Sprintf("Revoked access to %v", group),
	//                         "",
	//                         "",
	//                         dbc)
	//                 }

	//                 c.Groups[k] = value == "true"
	//                 list = append(list, kv{CardGroups.Append(gid), c.Groups[k]})
	//             }
	//         }
	//     }

	//     if strings.TrimSpace(c.Name) == "" && (c.Card == nil || *c.Card == 0) {
	//         if a != nil {
	//             if err := a.CanDelete(clone, auth.Cards); err != nil {
	//                 return nil, err
	//             }
	//         }

	//         if p := stringify(clone.Card, ""); p != "" {
	//             c.log(a,
	//                 "delete",
	//                 c.OID,
	//                 "card",
	//                 fmt.Sprintf("Deleted card %v", p),
	//                 "",
	//                 "",
	//                 dbc)
	//         } else if p = stringify(clone.Name, ""); p != "" {
	//             c.log(a,
	//                 "delete",
	//                 c.OID,
	//                 "card",
	//                 fmt.Sprintf("Deleted card for %v", p),
	//                 "",
	//                 "",
	//                 dbc)
	//         } else {
	//             c.log(a,
	//                 "delete",
	//                 c.OID,
	//                 "card",
	//                 "Deleted card",
	//                 "",
	//                 "",
	//                 dbc)
	//         }

	//         c.deleted = core.DateTimeNow()
	//         list = append(list, kv{CardDeleted, c.deleted})

	//         catalog.Delete(c.OID)
	//     }

	//     list = append(list, kv{CardStatus, c.status()})

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
