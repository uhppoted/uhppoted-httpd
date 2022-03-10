package cards

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
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

func (cc *Cards) AsObjects(auth auth.OpAuth) []schema.Object {
	guard.RLock()
	defer guard.RUnlock()

	objects := []schema.Object{}
	for _, card := range cc.cards {
		if card.IsValid() || card.IsDeleted() {
			catalog.Join(&objects, card.AsObjects(auth)...)
		}
	}

	return objects
}

func (cc *Cards) UpdateByOID(auth auth.OpAuth, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if cc == nil {
		return nil, nil
	}

	for k, c := range cc.cards {
		if c.OID.Contains(oid) {
			objects, err := c.set(auth, oid, value, dbc)
			if err == nil {
				cc.cards[k] = c
			}

			return objects, err
		}
	}

	objects := []schema.Object{}

	if oid == "<new>" {
		if c, err := cc.add(auth, Card{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' card")
		} else {
			c.log(auth,
				"add",
				c.OID,
				"card",
				"Added 'new' card",
				"",
				"",
				dbc)

			cc.cards[c.OID] = c
			catalog.Join(&objects, catalog.NewObject(c.OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(c.OID, CardCreated, c.created))
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

func (cc *Cards) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var c Card
		if err := c.deserialize(v); err == nil {
			if _, ok := cc.cards[c.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", c.Card, c.OID)
			}

			cc.cards[c.OID] = &c
		}
	}

	for _, v := range cc.cards {
		catalog.PutT(v, v.OID)
		catalog.PutV(v.OID, CardNumber, v.Card)
		catalog.PutV(v.OID, CardName, v.Name)
	}

	return nil
}

func (cc *Cards) Save() (json.RawMessage, error) {
	if err := validate(*cc); err != nil {
		return nil, err
	}

	if err := cc.scrub(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, c := range cc.cards {
		if c.IsValid() && !c.IsDeleted() {
			if record, err := c.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

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

func (cc *Cards) Validate() error {
	if cc != nil {
		return validate(*cc)
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

func (cc *Cards) Lookup(card uint32) *Card {
	if card != 0 {
		for _, c := range cc.cards {
			if c.Card != nil && uint32(*c.Card) == card {
				return c
			}
		}
	}

	return nil
}

func (cc *Cards) add(a auth.OpAuth, c Card) (*Card, error) {
	oid := catalog.NewCard()
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

	return card, nil
}

func validate(cc Cards) error {
	cards := map[uint32]string{}

	for _, c := range cc.cards {
		if c.IsDeleted() {
			continue
		}

		if c.OID == "" {
			return fmt.Errorf("Invalid card OID (%v)", c.OID)
		}

		if c.Card != nil {
			card := uint32(*c.Card)
			if id, ok := cards[card]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate card number (%v)", card),
					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", card, id, c.OID),
				}
			}

			cards[card] = string(c.OID)
		}
	}

	return nil
}

func (cc *Cards) scrub() error {
	return nil
}
