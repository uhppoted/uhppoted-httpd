package cards

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type CardHolders map[catalog.OID]*CardHolder

type CardHolder struct {
	OID    catalog.OID
	Name   *types.Name
	Card   *types.Card
	From   *types.Date
	To     *types.Date
	Groups map[string]bool

	Created time.Time `json:"-"`
}

type object catalog.Object

const CardCreated = catalog.CardCreated
const CardName = catalog.CardName
const CardNumber = catalog.CardNumber
const CardFrom = catalog.CardFrom
const CardTo = catalog.CardTo
const CardGroups = catalog.CardGroups

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
	}

	return replicant
}

func (c *CardHolder) IsValid() bool {
	return true
}

func (c *CardHolder) IsDeleted() bool {
	return false
}

func (c *CardHolder) AsObjects() []interface{} {
	status := stringify(types.StatusOk)
	created := c.Created.Format("2006-01-02 15:04:05")
	name := stringify(c.Name)
	number := stringify(c.Card)
	from := stringify(c.From)
	to := stringify(c.To)

	objects := []interface{}{
		object{OID: string(c.OID), Value: status},
		object{OID: catalog.Join(c.OID, CardCreated), Value: created},
		object{OID: catalog.Join(c.OID, CardName), Value: name},
		object{OID: catalog.Join(c.OID, CardNumber), Value: number},
		object{OID: catalog.Join(c.OID, CardFrom), Value: from},
		object{OID: catalog.Join(c.OID, CardTo), Value: to},
	}

	keys := []string{}
	for k, _ := range c.Groups {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for ix, k := range keys {
		v, ok := c.Groups[k]
		member := ok && v
		objects = append(objects, object{
			OID:   catalog.Join(c.OID, CardGroups.Append(fmt.Sprintf(".%v.1", ix+1))),
			Value: stringify(k),
		})

		objects = append(objects, object{
			OID:   catalog.Join(c.OID, CardGroups.Append(fmt.Sprintf(".%v.2", ix+1))),
			Value: stringify(member),
		})
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
		// name := stringify(c.Name)

		switch oid {
		case catalog.Join(c.OID, CardName):
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

		case catalog.Join(c.OID, CardNumber):
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

		case catalog.Join(c.OID, CardFrom):
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

		case catalog.Join(c.OID, CardTo):
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
		}
	}

	return objects, nil
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
