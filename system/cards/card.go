package cards

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Card struct {
	catalog.CatalogCard
	name   string
	card   uint32
	from   core.Date
	to     core.Date
	groups map[schema.OID]bool

	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

const BLANK = "'blank'"

var created = types.TimestampNow()

func (c Card) String() string {
	name := strings.TrimSpace(c.name)
	if name == "" {
		name = "-"
	}

	number := "-"

	if c.card != 0 {
		number = fmt.Sprintf("%v", c.card)
	}

	return fmt.Sprintf("%v (%v)", number, name)
}

func (c Card) AsAclCard() (core.Card, bool) {
	from := core.Date(c.from)
	to := core.Date(c.to)

	card := core.Card{
		CardNumber: c.card,
		From:       &from,
		To:         &to,
		Doors:      map[uint8]int{1: 0, 2: 0, 3: 0, 4: 0},
	}

	return card, c.card != 0 && c.from.IsValid() && c.to.IsValid()
}

func (c Card) CardNumber() uint32 {
	return c.card
}

func (c Card) Groups() []schema.OID {
	groups := []schema.OID{}

	for oid, member := range c.groups {
		if member {
			groups = append(groups, oid)
		}
	}

	return groups
}

func (c Card) IsValid() bool {
	return c.validate() == nil
}

func (c Card) validate() error {
	if strings.TrimSpace(c.name) == "" && c.card == 0 {
		return fmt.Errorf("At least one of card name and number must be defined")
	}

	return nil
}

func (c Card) IsDeleted() bool {
	return !c.deleted.IsZero()
}

func (c *Card) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if c.IsDeleted() {
		list = append(list, kv{CardDeleted, c.deleted})
	} else {
		name := c.name
		from := c.from
		to := c.to

		list = append(list, kv{CardStatus, c.Status()})
		list = append(list, kv{CardCreated, c.created})
		list = append(list, kv{CardDeleted, c.deleted})
		list = append(list, kv{CardName, name})
		list = append(list, kv{CardNumber, c.card})
		list = append(list, kv{CardFrom, from})
		list = append(list, kv{CardTo, to})

		groups := catalog.GetGroups()
		re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

		for _, group := range groups {
			g := group

			if m := re.FindStringSubmatch(string(g)); m != nil && len(m) > 2 {
				gid := m[2]
				member := c.groups[g]

				list = append(list, kv{CardGroups.Append(gid), member})
				list = append(list, kv{CardGroups.Append(gid + ".1"), group})
			}
		}
	}

	return c.toObjects(list, a)
}

func (c *Card) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Name   string
		Number uint32
		From   string
		To     string
		Groups []string
	}{}

	if c != nil {
		entity.Name = c.name
		entity.Number = c.card
		entity.From = fmt.Sprintf("%v", c.from)
		entity.To = fmt.Sprintf("%v", c.to)
		entity.Groups = []string{}

		for k, v := range c.groups {
			if v {
				if g := catalog.GetV(k, GroupName); g != nil {
					entity.Groups = append(entity.Groups, fmt.Sprintf("%v", g))
				}
			}
		}
	}

	return "card", &entity
}

func (c *Card) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if c == nil {
		return []schema.Object{}, nil
	}

	if c.IsDeleted() {
		return c.toObjects([]kv{{CardDeleted, c.deleted}}, a), fmt.Errorf("Card has been deleted")
	}

	f := func(field string, value interface{}) error {
		if a != nil {
			return a.CanUpdate(c, field, value, auth.Cards)
		}

		return nil
	}

	uid := auth.UID(a)
	list := []kv{}

	switch {
	case oid == c.OID.Append(CardName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			c.log(dbc, uid, "update", c.OID, "name", c.name, value, "Updated name from '%v' to '%v'", c.name, value)

			c.name = strings.TrimSpace(value)
			c.modified = types.TimestampNow()

			list = append(list, kv{CardName, stringify(c.name, "")})
		}

	case oid == c.OID.Append(CardNumber):
		if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
			if number, err := strconv.ParseUint(value, 10, 32); err != nil {
				return nil, err
			} else if err := f("number", number); err != nil {
				return nil, err
			} else {
				c.log(dbc, uid, "update", c.OID, "card", c.card, number, "Updated card number from %v to %v", c.card, value)

				c.card = uint32(number)
				c.modified = types.TimestampNow()

				list = append(list, kv{CardNumber, c.card})
			}
		} else if value == "" {
			if err := f("number", 0); err != nil {
				return nil, err
			} else {
				if p := stringify(c.name, ""); p != "" {
					c.log(dbc, uid, "update", c.OID, "number", c.card, p, "Cleared card number %v for %v", c.card, p)
				} else {
					c.log(dbc, uid, "update", c.OID, "number", c.card, "", "Cleared card number %v", c.card)
				}

				c.card = 0
				c.modified = types.TimestampNow()

				list = append(list, kv{CardNumber, ""})
			}
		}

	case oid == c.OID.Append(CardFrom):
		if err := f("from", value); err != nil {
			return nil, err
		} else if from, err := core.DateFromString(value); err != nil {
			return nil, err
		} else if !from.IsValid() {
			return nil, fmt.Errorf("invalid 'from' date (%v)", value)
		} else {
			c.log(dbc, uid, "update", c.OID, "from", c.from, value, "Updated VALID FROM date from %v to %v", c.from, value)

			c.from = from
			c.modified = types.TimestampNow()

			list = append(list, kv{CardFrom, c.from})
		}

	case oid == c.OID.Append(CardTo):
		if err := f("to", value); err != nil {
			return nil, err
		} else if to, err := core.DateFromString(value); err != nil {
			return nil, err
		} else if !to.IsValid() {
			return nil, fmt.Errorf("invalid 'to' date (%v)", value)
		} else {
			c.log(dbc, uid, "update", c.OID, "to", c.to, value, "Updated VALID UNTIL date from %v to %v", c.to, value)
			c.to = to
			c.modified = types.TimestampNow()

			list = append(list, kv{CardTo, c.to})
		}

	case schema.OID(c.OID.Append(CardGroups)).Contains(oid):
		if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
			gid := m[1]
			k := schema.GroupsOID.AppendS(gid)

			if err := f("group", value); err != nil {
				return nil, err
			} else if !catalog.HasGroup(schema.OID(k)) {
				return nil, fmt.Errorf("invalid group OID (%v)", k)
			} else {
				group := catalog.GetV(schema.OID(k), GroupName)

				if value == "true" {
					c.log(dbc, uid, "update", c.OID, "group", "", "", "Granted access to %v", group)
				} else {
					c.log(dbc, uid, "update", c.OID, "group", "", "", "Revoked access to %v", group)
				}

				c.groups[k] = value == "true"
				c.modified = types.TimestampNow()

				list = append(list, kv{CardGroups.Append(gid), c.groups[k]})
			}
		}
	}

	list = append(list, kv{CardStatus, c.Status()})

	return c.toObjects(list, a), nil
}

func (c *Card) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if c != nil {
		if a != nil {
			if err := a.CanDelete(c, auth.Cards); err != nil {
				return nil, err
			}
		}

		uid := auth.UID(a)
		if p := stringify(c.card, ""); p != "" {
			c.log(dbc, uid, "delete", c.OID, "card", "", "", "Deleted card %v", p)
		} else if p = stringify(c.name, ""); p != "" {
			c.log(dbc, uid, "delete", c.OID, "card", "", "", "Deleted card for %v", p)
		} else {
			c.log(dbc, uid, "delete", c.OID, "card", "", "", "Deleted card")
		}

		c.deleted = types.TimestampNow()
		c.modified = types.TimestampNow()

		list = append(list, kv{CardDeleted, c.deleted})
		list = append(list, kv{CardStatus, c.Status()})

		catalog.DeleteT(c.CatalogCard, c.OID)
	}

	return c.toObjects(list, a), nil
}

func (c *Card) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	f := func(c *Card, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(c, field, value, auth.Cards); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if !c.IsDeleted() && f(c, "OID", c.OID) {
		catalog.Join(&objects, catalog.NewObject(c.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(c, field, v.value) {
			catalog.Join(&objects, catalog.NewObject2(c.OID, v.field, v.value))
		}
	}

	return objects
}

func (c *Card) Status() types.Status {
	if c.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (c Card) serialize() ([]byte, error) {
	record := struct {
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		Card     types.Uint32    `json:"card,omitempty"`
		From     core.Date       `json:"from,omitempty"`
		To       core.Date       `json:"to,omitempty"`
		Groups   []schema.OID    `json:"groups"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
	}{
		OID:      c.OID,
		Name:     strings.TrimSpace(c.name),
		Card:     types.Uint32(c.card),
		From:     c.from,
		To:       c.to,
		Groups:   []schema.OID{},
		Created:  c.created.UTC(),
		Modified: c.modified.UTC(),
	}

	groups := catalog.GetGroups()

	for _, g := range groups {
		if c.groups[g] {
			record.Groups = append(record.Groups, g)
		}
	}

	return json.Marshal(record)
}

func (c *Card) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		Card     types.Uint32    `json:"card,omitempty"`
		From     core.Date       `json:"from,omitempty"`
		To       core.Date       `json:"to,omitempty"`
		Groups   []schema.OID    `json:"groups"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
	}{
		Groups:  []schema.OID{},
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.name = strings.TrimSpace(record.Name)
	c.card = uint32(record.Card)
	c.from = record.From
	c.to = record.To
	c.groups = map[schema.OID]bool{}
	c.created = record.Created
	c.modified = record.Modified

	for _, g := range record.Groups {
		c.groups[g] = true
	}

	return nil
}

func (c *Card) clone() *Card {
	var groups = map[schema.OID]bool{}

	for gid, g := range c.groups {
		groups[gid] = g
	}

	replicant := &Card{
		CatalogCard: catalog.CatalogCard{
			OID: c.OID,
		},
		name:   c.name,
		card:   c.card,
		from:   c.from,
		to:     c.to,
		groups: groups,

		created:  c.created,
		modified: c.modified,
		deleted:  c.deleted,
	}

	return replicant
}

func (c *Card) log(dbc db.DBC, uid, operation string, oid schema.OID, field string, before, after any, format string, fields ...any) {
	record := audit.AuditRecord{
		UID:       uid,
		OID:       oid,
		Component: "card",
		Operation: operation,
		Details: audit.Details{
			ID:          stringify(types.Uint32(c.card), ""),
			Name:        c.name,
			Field:       field,
			Description: fmt.Sprintf(format, fields...),
			Before:      stringify(before, ""),
			After:       stringify(after, ""),
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}

func stringify(i interface{}, defval string) string {
	s := ""

	switch v := i.(type) {
	case *uint32:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

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
