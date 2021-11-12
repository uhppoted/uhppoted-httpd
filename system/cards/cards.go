package cards

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Cards struct {
	Cards map[catalog.OID]*Card `json:"cardholders"`
	file  string
}

type result struct {
	Updated []interface{} `json:"updated"`
	Deleted []interface{} `json:"deleted"`
}

var guard sync.RWMutex

func NewCards() Cards {
	return Cards{
		Cards: map[catalog.OID]*Card{},
	}
}

func (cc *Cards) Load(file string) error {
	blob := struct {
		Cards []json.RawMessage `json:"cards"`
	}{
		Cards: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &blob); err != nil {
		return err
	}

	for _, v := range blob.Cards {
		var c Card
		if err := c.deserialize(v); err == nil {
			if _, ok := cc.Cards[c.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", c.Card, c.OID)
			}

			cc.Cards[c.OID] = &c
		}
	}

	for _, v := range cc.Cards {
		catalog.PutCard(v.OID)
		catalog.PutV(v.OID, catalog.CardNumber, v.Card)
		catalog.PutV(v.OID, catalog.CardName, v.Name)
	}

	cc.file = file

	return nil
}

func (cc *Cards) Save() error {
	if err := validate(*cc); err != nil {
		return err
	}

	if err := cc.scrub(); err != nil {
		return err
	}

	if cc.file == "" {
		return nil
	}

	serializable := struct {
		Cards []json.RawMessage `json:"cards"`
	}{
		Cards: []json.RawMessage{},
	}

	for _, c := range cc.Cards {
		if c.IsValid() && !c.IsDeleted() {
			if record, err := c.serialize(); err == nil && record != nil {
				serializable.Cards = append(serializable.Cards, record)
			}
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-cards.*")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cc.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), cc.file)
}

func (cc *Cards) Clone() Cards {
	shadow := Cards{
		Cards: map[catalog.OID]*Card{},
		file:  cc.file,
	}

	for cid, v := range cc.Cards {
		shadow.Cards[cid] = v.clone()
	}

	return shadow
}

func (cc *Cards) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if cc == nil {
		return nil, nil
	}

	for k, c := range cc.Cards {
		if c.OID.Contains(oid) {
			objects, err := c.set(auth, oid, value, dbc)
			if err == nil {
				cc.Cards[k] = c
			}

			return objects, err
		}
	}

	objects := []catalog.Object{}

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

			cc.Cards[c.OID] = c
			objects = append(objects, catalog.NewObject(c.OID, "new"))
			objects = append(objects, catalog.NewObject2(c.OID, CardCreated, c.created))
		}
	}

	return objects, nil
}

func (cc *Cards) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func (cc *Cards) Print() {
	if b, err := json.MarshalIndent(cc.Cards, "", "  "); err == nil {
		fmt.Printf("----------------- CARDS\n%s\n", string(b))
	}
}

func (cc *Cards) AsObjects() []interface{} {
	objects := []interface{}{}
	guard.RLock()

	defer guard.RUnlock()

	for _, card := range cc.Cards {
		if card.IsValid() || card.IsDeleted() {
			if l := card.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (cc *Cards) Lookup(card uint32) *Card {
	if card != 0 {
		for _, c := range cc.Cards {
			if c.Card != nil && uint32(*c.Card) == card {
				return c
			}
		}
	}

	return nil
}

func (cc *Cards) add(auth auth.OpAuth, c Card) (*Card, error) {
	oid := catalog.NewCard()
	if _, ok := cc.Cards[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	card := c.clone()
	card.OID = oid
	card.created = types.DateTime(time.Now())

	if auth != nil {
		if err := auth.CanAddCard(card); err != nil {
			return nil, err
		}
	}

	return card, nil
}

func validate(cc Cards) error {
	cards := map[uint32]string{}

	for _, c := range cc.Cards {
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
