package cards

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Card struct {
	OID    catalog.OID
	Name   *types.Name
	Card   *types.Card
	From   *types.Date
	To     *types.Date
	Groups map[catalog.OID]bool

	Created time.Time `json:"-"`
	deleted *time.Time
}

const CardCreated = catalog.CardCreated
const CardName = catalog.CardName
const CardNumber = catalog.CardNumber
const CardFrom = catalog.CardFrom
const CardTo = catalog.CardTo
const CardGroups = catalog.CardGroups
const GroupName = catalog.GroupName

var created = time.Now()

func (c Card) String() string {
	name := "-"
	number := "-"

	if c.Name != nil {
		name = fmt.Sprintf("%v", c.Name)
	}

	if c.Card != nil {
		number = fmt.Sprintf("%v", c.Card)
	}

	return fmt.Sprintf("%v (%v)", number, name)
}

func (c *Card) IsValid() bool {
	if c != nil {
		if c.Name != nil && *c.Name != "" {
			return true
		}

		if c.Card != nil && *c.Card != 0 {
			return true
		}
	}

	return false
}

func (c *Card) IsDeleted() bool {
	if c != nil && c.deleted != nil {
		return true
	}

	return false
}

func (c *Card) AsObjects() []interface{} {
	status := types.StatusOk
	created := c.Created.Format("2006-01-02 15:04:05")
	name := c.Name
	number := c.Card
	from := c.From
	to := c.To

	if c.deleted != nil {
		status = types.StatusDeleted
	}

	objects := []interface{}{
		catalog.NewObject(c.OID, status),
		catalog.NewObject2(c.OID, CardCreated, created),
		catalog.NewObject2(c.OID, CardName, name),
		catalog.NewObject2(c.OID, CardNumber, number),
		catalog.NewObject2(c.OID, CardFrom, from),
		catalog.NewObject2(c.OID, CardTo, to),
	}

	groups := catalog.Groups()
	re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

	for _, group := range groups {
		g := group

		if m := re.FindStringSubmatch(string(g)); m != nil && len(m) > 2 {
			gid := m[2]
			member := c.Groups[g]

			objects = append(objects, catalog.NewObject2(c.OID, CardGroups.Append(gid), member))
			objects = append(objects, catalog.NewObject2(c.OID, CardGroups.Append(gid+".1"), group))
		}
	}

	return objects
}

func (c *Card) AsRuleEntity() interface{} {
	type entity struct {
		Name   string
		Number uint32
		From   string
		To     string
		Groups []string
	}

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

		return &entity{
			Name:   name,
			Number: number,
			From:   from,
			To:     to,
			Groups: groups,
		}
	}

	return &entity{}
}

func (c *Card) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateCard(c, field, value)
	}

	if c != nil {
		clone := c.clone()

		switch {
		case oid == c.OID.Append(CardName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(c.Name, "<blank>"), stringify(value, "<blank>")), dbc)

				v := types.Name(value)
				c.Name = &v
				objects = append(objects, catalog.NewObject2(c.OID, CardName, c.Name))
			}

		case oid == c.OID.Append(CardNumber):
			if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if n, err := strconv.ParseUint(value, 10, 32); err != nil {
					return nil, err
				} else if err := f("number", n); err != nil {
					return nil, err
				} else {
					c.log(auth, "update", c.OID, "card", fmt.Sprintf("Updated card number from %v to %v", c.Card, value), dbc)
					v := types.Card(n)
					c.Card = &v
					objects = append(objects, catalog.NewObject2(c.OID, CardNumber, c.Card))
				}
			} else if value == "" {
				if err := f("number", 0); err != nil {
					return nil, err
				} else {
					if p := stringify(c.Name, ""); p != "" {
						c.log(auth, "update", c.OID, "number", fmt.Sprintf("Cleared card number %v for %v", c.Card, p), dbc)
					} else {
						c.log(auth, "update", c.OID, "number", fmt.Sprintf("Cleared card number %v", c.Card), dbc)
					}

					c.Card = nil
					objects = append(objects, catalog.NewObject2(c.OID, CardNumber, ""))
				}
			}

		case oid == c.OID.Append(CardFrom):
			if err := f("from", value); err != nil {
				return nil, err
			} else if from, err := types.ParseDate(value); err != nil {
				return nil, err
			} else if from == nil {
				return nil, fmt.Errorf("invalid 'from' date (%v)", value)
			} else {
				c.log(auth, "update", c.OID, "from", fmt.Sprintf("Updated VALID FROM date from %v to %v", c.From, value), dbc)
				c.From = from
				objects = append(objects, catalog.NewObject2(c.OID, CardFrom, c.From))
			}

		case oid == c.OID.Append(CardTo):
			if err := f("to", value); err != nil {
				return nil, err
			} else if to, err := types.ParseDate(value); err != nil {
				return nil, err
			} else if to == nil {
				return nil, fmt.Errorf("invalid 'to' date (%v)", value)
			} else {
				c.log(auth, "update", c.OID, "to", fmt.Sprintf("Updated VALID UNTIL date from %v to %v", c.From, value), dbc)
				c.To = to
				objects = append(objects, catalog.NewObject2(c.OID, CardTo, c.To))
			}

		case catalog.OID(c.OID.Append(CardGroups)).Contains(oid):
			if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
				gid := m[1]
				k := catalog.OID("0.4." + gid)

				if err := f("group", value); err != nil {
					return nil, err
				} else {
					group, _ := catalog.GetV(catalog.OID(k).Append(GroupName))

					if value == "true" {
						c.log(auth, "update", c.OID, "group", fmt.Sprintf("Granted access to %v", group), dbc)
					} else {
						c.log(auth, "update", c.OID, "group", fmt.Sprintf("Revoked access to %v", group), dbc)
					}

					c.Groups[k] = value == "true"
					objects = append(objects, catalog.NewObject2(c.OID, CardGroups.Append(gid), c.Groups[k]))
				}
			}
		}

		if (c.Name == nil || *c.Name == "") && (c.Card == nil || *c.Card == 0) {
			if auth != nil {
				if err := auth.CanDeleteCard(clone); err != nil {
					return nil, err
				}
			}

			if p := stringify(clone.Card, ""); p != "" {
				c.log(auth, "delete", c.OID, "card", fmt.Sprintf("Deleted card %v", p), dbc)
			} else if p = stringify(clone.Name, ""); p != "" {
				c.log(auth, "delete", c.OID, "card", fmt.Sprintf("Deleted card for %v", p), dbc)
			} else {
				c.log(auth, "delete", c.OID, "card", "Deleted card", dbc)
			}

			now := time.Now()
			c.deleted = &now

			objects = append(objects, catalog.NewObject(c.OID, "deleted"))

			catalog.Delete(c.OID)
		}

	}

	return objects, nil
}

func (c Card) GetName() string {
	if c.Name != nil {
		return string(*c.Name)
	}

	return ""
}

func (c Card) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		Card    uint32        `json:"card,omitempty"`
		From    *types.Date   `json:"from,omitempty"`
		To      *types.Date   `json:"to,omitempty"`
		Groups  []catalog.OID `json:"groups"`
		Created string        `json:"created"`
	}{
		OID:     c.OID,
		From:    c.From,
		To:      c.To,
		Groups:  []catalog.OID{},
		Created: c.Created.Format("2006-01-02 15:04:05"),
	}

	if c.Name != nil {
		record.Name = fmt.Sprintf("%v", c.Name)
	}

	if c.Card != nil {
		record.Card = uint32(*c.Card)
	}

	groups := catalog.Groups()

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
		From    *types.Date   `json:"from,omitempty"`
		To      *types.Date   `json:"to,omitempty"`
		Groups  []catalog.OID `json:"groups"`
		Created string        `json:"created"`
	}{
		Groups:  []catalog.OID{},
		Created: created.Format("2006-01-02 15:04:05"),
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	c.OID = record.OID
	c.From = record.From
	c.To = record.To
	c.Groups = map[catalog.OID]bool{}

	if record.Name != "" {
		c.Name = (*types.Name)(&record.Name)
	}

	if record.Card != 0 {
		c.Card = (*types.Card)(&record.Card)
	}

	for _, g := range record.Groups {
		c.Groups[g] = true
	}

	if t, err := time.Parse("2006-01-02 15:04:05", record.Created); err == nil {
		c.Created = t
	}

	return nil
}

func (c *Card) clone() *Card {
	name := c.Name.Copy()
	card := c.Card.Copy()
	var groups = map[catalog.OID]bool{}

	for gid, g := range c.Groups {
		groups[gid] = g
	}

	replicant := &Card{
		OID:    c.OID,
		Name:   name,
		Card:   card,
		From:   c.From,
		To:     c.To,
		Groups: groups,

		Created: c.Created,
		deleted: c.deleted,
	}

	return replicant
}

func (c Card) stash() {
}

func (c *Card) log(auth auth.OpAuth, operation string, oid catalog.OID, field string, description string, dbc db.DBC) {
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
		},
	}

	dbc.Write(record)
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
		if i != nil {
			s = fmt.Sprintf("%v", i)
		}
	}

	if s != "" {
		return s
	}

	return defval
}
