package cards

import (
	"fmt"
	"sort"
	"time"

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

// TODO make unexported after rationalising 'Cards' implementation
func (c *CardHolder) Set(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		//			if auth == nil {
		//				return nil
		//			}
		//
		//			return auth.CanUpdateDoor(d, field, value)
		return nil
	}

	if c != nil {
		// name := stringify(c.Name)

		switch oid {
		case catalog.Join(c.OID, CardName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				//				d.log(auth, "update", d.OID, "name", stringify(d.Name), value)
				v := types.Name(value)
				c.Name = &v
				objects = append(objects, object{
					OID:   c.OID.Append(CardName),
					Value: stringify(c.Name),
				})
			}

			//		case d.OID + DoorDelay:
			//			delay := d.delay
			//
			//			if err := f("delay", value); err != nil {
			//				return nil, err
			//			} else if v, err := strconv.ParseUint(value, 10, 8); err != nil {
			//				return nil, err
			//			} else {
			//				d.delay = uint8(v)
			//
			//				catalog.PutV(d.OID+DoorDelayConfigured, d.delay, true)
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorDelay,
			//					Value: stringify(d.delay),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorDelayStatus,
			//					Value: stringify(types.StatusUncertain),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorDelayConfigured,
			//					Value: stringify(d.delay),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorDelayError,
			//					Value: "",
			//				})
			//
			//				d.log(auth, "update", d.OID, "delay", stringify(delay), value)
			//			}
			//
			//		case d.OID + DoorControl:
			//			if err := f("mode", value); err != nil {
			//				return nil, err
			//			} else {
			//				mode := d.mode
			//				switch value {
			//				case "controlled":
			//					d.mode = core.Controlled
			//				case "normally open":
			//					d.mode = core.NormallyOpen
			//				case "normally closed":
			//					d.mode = core.NormallyClosed
			//				default:
			//					return nil, fmt.Errorf("%v: invalid control state (%v)", d.Name, value)
			//				}
			//
			//				catalog.PutV(d.OID+DoorControlConfigured, d.mode, true)
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorControl,
			//					Value: stringify(d.mode),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorControlStatus,
			//					Value: stringify(types.StatusUncertain),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorControlConfigured,
			//					Value: stringify(d.mode),
			//				})
			//
			//				objects = append(objects, object{
			//					OID:   d.OID + DoorControlError,
			//					Value: "",
			//				})
			//
			//				d.log(auth, "update", d.OID, "mode", stringify(mode), value)
			//			}
			//		}
			//
			//		if !d.IsValid() {
			//			if auth != nil {
			//				if err := auth.CanDeleteDoor(d); err != nil {
			//					return nil, err
			//				}
			//			}
			//
			//			d.log(auth, "delete", d.OID, "name", name, "")
			//			now := time.Now()
			//			d.deleted = &now
			//
			//			objects = append(objects, object{
			//				OID:   d.OID,
			//				Value: "deleted",
			//			})
			//
			//			catalog.Delete(d.OID)
		}
	}

	return objects, nil
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
