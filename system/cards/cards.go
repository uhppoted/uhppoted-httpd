package cards

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Cards struct {
	cards map[schema.OID]*Card
}

var guard sync.RWMutex

func NewCards() Cards {
	return Cards{
		cards: map[schema.OID]*Card{},
	}
}

func (cc Cards) Lookup(card uint32) (*Card, bool) {
	guard.RLock()
	defer guard.RUnlock()

	if card != 0 {
		for _, v := range cc.cards {
			if v.CardID == card {
				return v, v.unconfigured
			}
		}
	}

	return nil, false
}

func (cc *Cards) AsObjects(a *auth.Authorizator) []schema.Object {
	guard.RLock()
	defer guard.RUnlock()

	objects := []schema.Object{}
	for _, card := range cc.cards {
		if card.IsValid() || card.IsDeleted() {
			catalog.Join(&objects, card.AsObjects(a)...)
		}
	}

	return objects
}

func (cc *Cards) UpdateByOID(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if cc != nil {
		for k, c := range cc.cards {
			if c.OID.Contains(oid) {
				objects, err := c.set(a, oid, value, dbc)
				if err == nil {
					cc.cards[k] = c
				}

				return objects, err
			}
		}

		if oid == "<new>" {
			if c, err := cc.add(a, Card{}); err != nil {
				return nil, err
			} else if c == nil {
				return nil, fmt.Errorf("Failed to add 'new' card")
			} else {
				c.log(dbc, auth.UID(a), "add", "card", "", "", "Added 'new' card")

				catalog.Join(&objects, catalog.NewObject(c.OID, "new"))
				catalog.Join(&objects, catalog.NewObject2(c.OID, CardCreated, c.created))
			}
		}
	}

	return objects, nil
}

func (cc *Cards) DeleteByOID(a *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if cc != nil {
		for k, c := range cc.cards {
			if c.OID == oid {
				objects, err := c.delete(a, dbc)
				if err == nil {
					cc.cards[k] = c
				}

				return objects, err
			}
		}
	}

	return objects, nil
}

func (cc *Cards) List() []Card {
	list := []Card{}

	for _, c := range cc.cards {
		if c != nil {
			list = append(list, *c)
		}
	}

	return list
}

func (cc *Cards) Found(list []uint32) {
	if cc == nil {
		return
	}

	guard.Lock()
	defer guard.Unlock()

	remove := []*Card{}
	add := make([]uint32, len(list))

	copy(add, list)

loop:
	for _, card := range cc.cards {
		for ix, c := range add {
			if card.CardID == c {
				add[ix] = add[len(add)-1]
				add = add[:len(add)-1]

				continue loop
			}
		}

		if card.unconfigured {
			remove = append(remove, card)
		}
	}

	for _, card := range remove {
		delete(cc.cards, card.OID)
	}

	for _, c := range add {
		card := Card{
			CatalogCard: catalog.CatalogCard{
				CardID: c,
			},
		}

		oid := catalog.NewT(card.CatalogCard)
		if _, ok := cc.cards[oid]; ok {
			log.Warnf("Duplicate catalog entry (%v) for unconfigured card %v", oid, c)
		} else {
			card.OID = oid
			card.created = types.TimestampNow()
			card.unconfigured = true

			cc.cards[card.OID] = &card

			log.Infof("Adding unconfigured card %v", c)
		}
	}
}

func (cc *Cards) MarkIncorrect(list []uint32) {
	if cc != nil {
	loop:
		for _, card := range cc.cards {
			for _, c := range list {
				if c == card.CardID {
					card.incorrect = true
					continue loop
				}
			}

			card.incorrect = false
		}

		for _, c := range cc.cards {
			catalog.PutV(c.OID, CardStatus, c.Status())
		}
	}
}

func (cc *Cards) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var c Card
		if err := c.deserialize(v); err == nil {
			if _, ok := cc.cards[c.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", c.CardID, c.OID)
			}

			cc.cards[c.OID] = &c
		}
	}

	for _, c := range cc.cards {
		catalog.PutT(c.CatalogCard)
		catalog.PutV(c.OID, CardNumber, c.CardID)
		catalog.PutV(c.OID, CardName, c.name)
	}

	return nil
}

func (cc *Cards) Save() (json.RawMessage, error) {
	if err := cc.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, c := range cc.cards {
		if c.IsValid() && !c.IsDeleted() && !c.unconfigured {
			if record, err := c.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

// NTS: 'added' is specifically not cloned - it has a lifetime for the duration of
//      the 'shadow' copy only
func (cc *Cards) Clone() Cards {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Cards{
		cards: map[schema.OID]*Card{},
	}

	for cid, v := range cc.cards {
		shadow.cards[cid] = v.clone()
	}

	return shadow
}

func (cc Cards) Validate() error {
	cards := map[uint32]string{}

	for k, c := range cc.cards {
		if c.IsDeleted() {
			continue
		}

		if c.OID == "" {
			return fmt.Errorf("Invalid card OID (%v)", c.OID)
		} else if k != c.OID {
			return fmt.Errorf("Card %s: mismatched OID %v (expected %v)", c.name, c.OID, k)
		}

		if err := c.validate(); err != nil {
			if !c.modified.IsZero() {
				return err
			}
		}

		if c.CardID != 0 {
			if id, ok := cards[c.CardID]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate card number (%v)", c.CardID),
					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", c.CardID, id, c.OID),
				}
			}

			cards[c.CardID] = string(c.OID)
		}
	}

	return nil
}

func (cc *Cards) Print() {
	serializable := []json.RawMessage{}
	for _, c := range cc.cards {
		if c.IsValid() && !c.IsDeleted() {
			if record, err := c.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- CARDS\n%s\n", string(b))
	}
}

func (cc *Cards) Sweep(retention time.Duration) {
	if cc != nil {
		cutoff := time.Now().Add(-retention)
		for i, v := range cc.cards {
			if v.IsDeleted() && v.deleted.Before(cutoff) {
				delete(cc.cards, i)
			}
		}
	}
}

func (cc *Cards) add(a *auth.Authorizator, c Card) (*Card, error) {
	oid := catalog.NewT(c.CatalogCard)
	if _, ok := cc.cards[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	card := c.clone()
	card.OID = oid
	card.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(card, auth.Cards); err != nil {
			return nil, err
		}
	}

	cc.cards[card.OID] = card

	return card, nil
}
