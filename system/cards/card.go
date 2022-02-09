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
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Card struct {
	OID    catalog.OID
	Name   string
	Card   *types.Card
	From   core.Date
	To     core.Date
	Groups map[catalog.OID]bool

	created core.DateTime
	deleted core.DateTime
}

type kv = struct {
	field catalog.Suffix
	value interface{}
}

const BLANK = "'blank'"

var created = core.DateTimeNow()

func (c Card) String() string {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		name = "-"
	}

	number := "-"

	if c.Card != nil {
		number = fmt.Sprintf("%v", c.Card)
	}

	return fmt.Sprintf("%v (%v)", number, name)
}

func (c Card) GetName() string {
	return strings.TrimSpace(c.Name)
}

func (c *Card) IsValid() bool {
	if c != nil {
		if strings.TrimSpace(c.Name) != "" {
			return true
		}

		if c.Card != nil && *c.Card != 0 {
			return true
		}
	}

	return false
}

func (c Card) IsDeleted() bool {
	return !c.deleted.IsZero()
}

func (c *Card) AsObjects(auth auth.OpAuth) []catalog.Object {
	list := []kv{}

	if c.IsDeleted() {
		list = append(list, kv{CardDeleted, c.deleted})
	} else {
		name := c.Name
		number := c.Card
		from := c.From
		to := c.To

		list = append(list, kv{CardStatus, c.status()})
		list = append(list, kv{CardCreated, c.created})
		list = append(list, kv{CardDeleted, c.deleted})
		list = append(list, kv{CardName, name})
		list = append(list, kv{CardNumber, number})
		list = append(list, kv{CardFrom, from})
		list = append(list, kv{CardTo, to})

		groups := catalog.GetGroups()
		re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

		for _, group := range groups {
			g := group

			if m := re.FindStringSubmatch(string(g)); m != nil && len(m) > 2 {
				gid := m[2]
				member := c.Groups[g]

				list = append(list, kv{CardGroups.Append(gid), member})
				list = append(list, kv{CardGroups.Append(gid + ".1"), group})
			}
		}
	}

	return c.toObjects(list, auth)
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
		name := fmt.Sprintf("%v", c.Name)
		number := uint32(0)
		from := fmt.Sprintf("%v", c.From)
		to := fmt.Sprintf("%v", c.To)

		if c.Card != nil {
			number = uint32(*c.Card)
		}

		groups := []string{}
		for k, v := range c.Groups {
			if v {
				groups = append(groups, string(k))
			}
		}

		entity.Name = name
		entity.Number = number
		entity.From = from
		entity.To = to
		entity.Groups = groups
	}

	return "card", &entity
}

func (c *Card) set(a auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if c == nil {
		return []catalog.Object{}, nil
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

	list := []kv{}
	clone := c.clone()

	switch {
	case oid == c.OID.Append(CardName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			c.log(a,
				"update",
				c.OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(c.Name, BLANK), stringify(value, BLANK)),
				stringify(c.Name, ""),
				stringify(value, ""),
				dbc)

			c.Name = strings.TrimSpace(value)
			list = append(list, kv{CardName, stringify(c.Name, "")})
		}

	case oid == c.OID.Append(CardNumber):
		if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
			if n, err := strconv.ParseUint(value, 10, 32); err != nil {
				return nil, err
			} else if err := f("number", n); err != nil {
				return nil, err
			} else {
				c.log(a,
					"update",
					c.OID,
					"card",
					fmt.Sprintf("Updated card number from %v to %v", c.Card, value),
					stringify(c.Card, ""),
					stringify(value, ""),
					dbc)

				v := types.Card(n)
				c.Card = &v
				list = append(list, kv{CardNumber, c.Card})
			}
		} else if value == "" {
			if err := f("number", 0); err != nil {
				return nil, err
			} else {
				if p := stringify(c.Name, ""); p != "" {
					c.log(a,
						"update",
						c.OID,
						"number",
						fmt.Sprintf("Cleared card number %v for %v", c.Card, p),
						stringify(c.Card, ""),
						stringify(p, ""),
						dbc)
				} else {
					c.log(a,
						"update",
						c.OID,
						"number",
						fmt.Sprintf("Cleared card number %v", c.Card),
						stringify(c.Card, ""),
						"",
						dbc)
				}

				c.Card = nil
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
			c.log(a,
				"update",
				c.OID,
				"from",
				fmt.Sprintf("Updated VALID FROM date from %v to %v", c.From, value),
				stringify(c.From, ""),
				stringify(value, ""),
				dbc)

			c.From = from
			list = append(list, kv{CardFrom, c.From})
		}

	case oid == c.OID.Append(CardTo):
		if err := f("to", value); err != nil {
			return nil, err
		} else if to, err := core.DateFromString(value); err != nil {
			return nil, err
		} else if to.IsValid() {
			return nil, fmt.Errorf("invalid 'to' date (%v)", value)
		} else {
			c.log(a,
				"update",
				c.OID,
				"to",
				fmt.Sprintf("Updated VALID UNTIL date from %v to %v", c.From, value),
				stringify(c.From, ""),
				stringify(value, ""),
				dbc)

			c.To = to
			list = append(list, kv{CardTo, c.To})
		}

	case catalog.OID(c.OID.Append(CardGroups)).Contains(oid):
		if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
			gid := m[1]
			k := catalog.GroupsOID.AppendS(gid)

			if err := f("group", value); err != nil {
				return nil, err
			} else if !catalog.HasGroup(catalog.OID(k)) {
				return nil, fmt.Errorf("invalid group OID (%v)", k)
			} else {
				group := catalog.GetV(catalog.OID(k), GroupName)

				if value == "true" {
					c.log(a,
						"update",
						c.OID,
						"group",
						fmt.Sprintf("Granted access to %v", group),
						"",
						"",
						dbc)
				} else {
					c.log(a,
						"update",
						c.OID,
						"group",
						fmt.Sprintf("Revoked access to %v", group),
						"",
						"",
						dbc)
				}

				c.Groups[k] = value == "true"
				list = append(list, kv{CardGroups.Append(gid), c.Groups[k]})
			}
		}
	}

	if strings.TrimSpace(c.Name) == "" && (c.Card == nil || *c.Card == 0) {
		if a != nil {
			if err := a.CanDelete(clone, auth.Cards); err != nil {
				return nil, err
			}
		}

		if p := stringify(clone.Card, ""); p != "" {
			c.log(a,
				"delete",
				c.OID,
				"card",
				fmt.Sprintf("Deleted card %v", p),
				"",
				"",
				dbc)
		} else if p = stringify(clone.Name, ""); p != "" {
			c.log(a,
				"delete",
				c.OID,
				"card",
				fmt.Sprintf("Deleted card for %v", p),
				"",
				"",
				dbc)
		} else {
			c.log(a,
				"delete",
				c.OID,
				"card",
				"Deleted card",
				"",
				"",
				dbc)
		}

		c.deleted = core.DateTimeNow()
		list = append(list, kv{CardDeleted, c.deleted})

		catalog.Delete(c.OID)
	}

	list = append(list, kv{CardStatus, c.status()})

	return c.toObjects(list, a), nil
}

func (c *Card) toObjects(list []kv, a auth.OpAuth) []catalog.Object {
	f := func(c *Card, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(c, field, value, auth.Cards); err != nil {
				return false
			}
		}

		return true
	}

	objects := []catalog.Object{}

	if !c.IsDeleted() && f(c, "OID", c.OID) {
		objects = append(objects, catalog.NewObject(c.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(c, field, v.value) {
			objects = append(objects, catalog.NewObject2(c.OID, v.field, v.value))
		}
	}

	return objects
}

func (c *Card) status() types.Status {
	if c.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (c Card) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		Card    uint32        `json:"card,omitempty"`
		From    core.Date     `json:"from,omitempty"`
		To      core.Date     `json:"to,omitempty"`
		Groups  []catalog.OID `json:"groups"`
		Created core.DateTime `json:"created"`
	}{
		OID:     c.OID,
		Name:    strings.TrimSpace(c.Name),
		From:    c.From,
		To:      c.To,
		Groups:  []catalog.OID{},
		Created: c.created,
	}

	if c.Card != nil {
		record.Card = uint32(*c.Card)
	}

	groups := catalog.GetGroups()

	for _, g := range groups {
		if c.Groups[g] {
			record.Groups = append(record.Groups, g)
		}
	}

	return json.Marshal(record)
}

func (c *Card) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		Card    uint32        `json:"card,omitempty"`
		From    core.Date     `json:"from,omitempty"`
		To      core.Date     `json:"to,omitempty"`
		Groups  []catalog.OID `json:"groups"`
		Created core.DateTime `json:"created"`
	}{
		Groups:  []catalog.OID{},
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.Name = strings.TrimSpace(record.Name)
	c.From = record.From
	c.To = record.To
	c.Groups = map[catalog.OID]bool{}
	c.created = record.Created

	if record.Card != 0 {
		c.Card = (*types.Card)(&record.Card)
	}

	for _, g := range record.Groups {
		c.Groups[g] = true
	}

	return nil
}

func (c *Card) clone() *Card {
	card := c.Card.Copy()
	var groups = map[catalog.OID]bool{}

	for gid, g := range c.Groups {
		groups[gid] = g
	}

	replicant := &Card{
		OID:    c.OID,
		Name:   c.Name,
		Card:   card,
		From:   c.From,
		To:     c.To,
		Groups: groups,

		created: c.created,
		deleted: c.deleted,
	}

	return replicant
}

func (c *Card) log(auth auth.OpAuth, operation string, oid catalog.OID, field, description, before, after string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       oid,
		Component: "card",
		Operation: operation,
		Details: audit.Details{
			ID:          stringify(c.Card, ""),
			Name:        stringify(c.Name, ""),
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
