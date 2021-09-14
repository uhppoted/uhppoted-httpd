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
	"github.com/uhppoted/uhppoted-httpd/types"
)

type CardHolder struct {
	OID    catalog.OID
	Name   *types.Name
	Card   *types.Card
	From   *types.Date
	To     *types.Date
	Groups map[string]bool

	Created time.Time `json:"-"`
	deleted *time.Time
}

type object catalog.Object

const CardCreated = catalog.CardCreated
const CardName = catalog.CardName
const CardNumber = catalog.CardNumber
const CardFrom = catalog.CardFrom
const CardTo = catalog.CardTo
const CardGroups = catalog.CardGroups

var created = time.Now()
var trail audit.Trail

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func (c CardHolder) String() string {
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

func (c *CardHolder) Clone() *CardHolder {
	name := c.Name.Copy()
	card := c.Card.Copy()
	var groups = map[string]bool{}

	for gid, g := range c.Groups {
		groups[gid] = g
	}

	replicant := &CardHolder{
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

func (c *CardHolder) IsValid() bool {
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

func (c *CardHolder) IsDeleted() bool {
	if c != nil && c.deleted != nil {
		return true
	}

	return false
}

func (c *CardHolder) AsObjects() []interface{} {
	status := stringify(types.StatusOk)
	created := c.Created.Format("2006-01-02 15:04:05")
	name := stringify(c.Name)
	number := stringify(c.Card)
	from := stringify(c.From)
	to := stringify(c.To)

	if c.deleted != nil {
		status = stringify(types.StatusDeleted)
	}

	objects := []interface{}{
		object{OID: string(c.OID), Value: status},
		object{OID: catalog.Join(c.OID, CardCreated), Value: created},
		object{OID: catalog.Join(c.OID, CardName), Value: name},
		object{OID: catalog.Join(c.OID, CardNumber), Value: number},
		object{OID: catalog.Join(c.OID, CardFrom), Value: from},
		object{OID: catalog.Join(c.OID, CardTo), Value: to},
	}

	groups := catalog.Groups()
	re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

	for _, group := range groups {
		g := fmt.Sprintf("%v", group)

		if m := re.FindStringSubmatch(g); m != nil && len(m) > 2 {
			gid := m[2]
			member := c.Groups[g]

			objects = append(objects, object{
				OID:   catalog.Join(c.OID, CardGroups.Append(gid)),
				Value: stringify(member),
			})

			objects = append(objects, object{
				OID:   catalog.Join(c.OID, CardGroups.Append(gid+".1")),
				Value: stringify(group),
			})
		}
	}

	return objects
}

func (c *CardHolder) AsRuleEntity() interface{} {
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
				groups = append(groups, k)
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

// FIXME make unexported after moving memdb implementation to this package
func (c CardHolder) Serialize() ([]byte, error) {
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
		if c.Groups[string(g)] {
			record.Groups = append(record.Groups, catalog.OID(g))
		}
	}

	return json.Marshal(record)
}

// FIXME make unexported after moving memdb implementation to this package
func (c *CardHolder) Deserialize(bytes []byte) error {
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
	c.Groups = map[string]bool{}

	if record.Name != "" {
		c.Name = (*types.Name)(&record.Name)
	}

	if record.Card != 0 {
		c.Card = (*types.Card)(&record.Card)
	}

	for _, g := range record.Groups {
		c.Groups[string(g)] = true
	}

	if t, err := time.Parse("2006-01-02 15:04:05", record.Created); err == nil {
		c.Created = t
	}

	return nil
}

// TODO make unexported after rationalising 'Cards' implementation
func (c *CardHolder) Set(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateCard(c, field, value)
	}

	if c != nil {
		clone := c.Clone()

		switch {
		case oid == catalog.Join(c.OID, CardName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				c.log(auth, "update", c.OID, "name", stringify(c.Name), value)
				v := types.Name(value)
				c.Name = &v
				objects = append(objects, object{
					OID:   c.OID.Append(CardName),
					Value: stringify(c.Name),
				})
			}

		case oid == catalog.Join(c.OID, CardNumber):
			if ok, err := regexp.MatchString("[0-9]+", value); err == nil && ok {
				if n, err := strconv.ParseUint(value, 10, 32); err != nil {
					return nil, err
				} else if err := f("number", n); err != nil {
					return nil, err
				} else {
					c.log(auth, "update", c.OID, "number", stringify(c.Card), value)
					v := types.Card(n)
					c.Card = &v
					objects = append(objects, object{
						OID:   c.OID.Append(CardNumber),
						Value: stringify(c.Card),
					})
				}
			} else if value == "" {
				if err := f("number", 0); err != nil {
					return nil, err
				} else {
					c.log(auth, "update", c.OID, "number", stringify(c.Card), value)
					c.Card = nil
					objects = append(objects, object{
						OID:   c.OID.Append(CardNumber),
						Value: "",
					})
				}
			}

		case oid == catalog.Join(c.OID, CardFrom):
			if err := f("from", value); err != nil {
				return nil, err
			} else if from, err := types.ParseDate(value); err != nil {
				return nil, err
			} else if from == nil {
				return nil, fmt.Errorf("invalid 'from' date (%v)", value)
			} else {
				c.log(auth, "update", c.OID, "from", stringify(c.From), value)
				c.From = from
				objects = append(objects, object{
					OID:   c.OID.Append(CardFrom),
					Value: stringify(c.From),
				})
			}

		case oid == catalog.Join(c.OID, CardTo):
			if err := f("to", value); err != nil {
				return nil, err
			} else if to, err := types.ParseDate(value); err != nil {
				return nil, err
			} else if to == nil {
				return nil, fmt.Errorf("invalid 'to' date (%v)", value)
			} else {
				c.log(auth, "update", c.OID, "to", stringify(c.To), value)
				c.To = to
				objects = append(objects, object{
					OID:   c.OID.Append(CardTo),
					Value: stringify(c.To),
				})
			}

		case catalog.OID(c.OID.Append(CardGroups)).Contains(oid):
			if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(oid); m != nil && len(m) > 1 {
				gid := m[1]
				k := "0.4." + gid

				if err := f("group", value); err != nil {
					return nil, err
				} else {
					c.log(auth, "update", c.OID, "group", k, value)
					c.Groups[k] = value == "true"
					objects = append(objects, object{
						OID:   catalog.Join(c.OID, CardGroups.Append(gid)),
						Value: stringify(c.Groups[gid]),
					})
				}
			}
		}

		if (c.Name == nil || *c.Name == "") && (c.Card == nil || *c.Card == 0) {
			if auth != nil {
				if err := auth.CanDeleteCard(clone); err != nil {
					return nil, err
				}
			}

			c.log(auth, "delete", c.OID, "card", "", "")
			now := time.Now()
			c.deleted = &now

			objects = append(objects, object{
				OID:   stringify(c.OID),
				Value: "deleted",
			})

			catalog.Delete(stringify(c.OID))
		}

	}

	return objects, nil
}

// TODO remove - temporary implementation pending memdb move to 'cards' package
func (c *CardHolder) Log(auth auth.OpAuth, operation string, oid catalog.OID, field, current, value string) {
	c.log(auth, operation, oid, field, current, value)
}

func (c *CardHolder) log(auth auth.OpAuth, operation string, oid catalog.OID, field, current, value string) {
	type info struct {
		OID     string `json:"OID"`
		Card    string `json:"card"`
		Field   string `json:"field"`
		Current string `json:"current"`
		Updated string `json:"new"`
	}

	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	if trail != nil {
		record := audit.LogEntry{
			UID:       uid,
			Module:    stringify(oid),
			Operation: operation,
			Info: info{
				OID:     stringify(oid),
				Card:    stringify(c.Card),
				Field:   field,
				Current: current,
				Updated: value,
			},
		}

		trail.Write(record)
	}
}

func lookup(oid string) interface{} {
	if v, _ := catalog.GetV(oid); v != nil {
		return v
	}

	return nil
}

func stringify(i interface{}) string {
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	default:
		if i != nil {
			return fmt.Sprintf("%v", i)
		}
	}

	return ""
}
