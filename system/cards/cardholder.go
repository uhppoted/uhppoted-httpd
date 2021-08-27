package cards

import (
	"fmt"
	"sort"

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
}

type object catalog.Object

const CardName = catalog.CardName
const CardNumber = catalog.CardNumber
const CardFrom = catalog.CardFrom
const CardTo = catalog.CardTo
const CardGroups = catalog.CardGroups

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
	name := stringify(c.Name)
	number := stringify(c.Card)
	from := stringify(c.From)
	to := stringify(c.To)

	objects := []interface{}{
		object{OID: string(c.OID), Value: status},
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
		Card   uint32
		Groups []string
	}

	if c != nil {
		cardNumber := uint32(0)
		if c.Card != nil {
			cardNumber = uint32(*c.Card)
		}

		groups := []string{}
		for k, v := range c.Groups {
			if v {
				groups = append(groups, k)
			}
		}

		return &entity{
			Name:   fmt.Sprintf("%v", c.Name),
			Card:   cardNumber,
			Groups: groups,
		}
	}

	return &entity{}
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
