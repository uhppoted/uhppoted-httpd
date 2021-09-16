package cards

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type fdb struct {
	Cards map[catalog.OID]*CardHolder `json:"cardholders"`
	file  string
}

type result struct {
	Updated []interface{} `json:"updated"`
	Deleted []interface{} `json:"deleted"`
}

const GroupName = catalog.GroupName

var guard sync.RWMutex
var trail audit.Trail

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func NewCards() *fdb {
	return &fdb{
		Cards: map[catalog.OID]*CardHolder{},
	}
}

func (cc *fdb) Load(file string) error {
	blob := struct {
		Cards []json.RawMessage `json:"cards"`
	}{
		Cards: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	for _, v := range blob.Cards {
		var c CardHolder
		if err := c.deserialize(v); err == nil {
			if _, ok := cc.Cards[c.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", c.Card, c.OID)
			}

			cc.Cards[c.OID] = &c
		}
	}

	for _, v := range cc.Cards {
		catalog.PutCard(v.OID)
	}

	cc.file = file

	return nil
}

func (cc *fdb) Save() error {
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

func (d *fdb) Clone() Cards {
	shadow := fdb{
		Cards: map[catalog.OID]*CardHolder{},
		file:  d.file,
	}

	for cid, v := range d.Cards {
		shadow.Cards[cid] = v.clone()
	}

	return &shadow
}

func (cc *fdb) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if cc == nil {
		return nil, nil
	}

	for k, c := range cc.Cards {
		if c.OID.Contains(oid) {
			objects, err := c.set(auth, oid, value)
			if err == nil {
				cc.Cards[k] = c
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	if oid == "<new>" {
		if c, err := cc.add(auth, CardHolder{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' card")
		} else {
			c.log(auth, "add", c.OID, "card", "", "")
			cc.Cards[c.OID] = c
			objects = append(objects, object{
				OID:   fmt.Sprintf("%v", c.OID),
				Value: "new",
			})
		}
	}

	return objects, nil
}

func (cc *fdb) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func (cc *fdb) Print() {
	if b, err := json.MarshalIndent(cc.Cards, "", "  "); err == nil {
		fmt.Printf("----------------- CARDS\n%s\n", string(b))
	}
}

func (cc *fdb) AsObjects() []interface{} {
	objects := []interface{}{}

	guard.RLock()

	defer guard.RUnlock()

	for _, record := range cc.Cards {
		if record.IsValid() || record.IsDeleted() {
			if l := record.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (cc *fdb) ACL(rules IRules) ([]types.Permissions, error) {
	guard.RLock()

	defer guard.RUnlock()

	list := []types.Permissions{}

	for _, c := range cc.Cards {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			var doors = []string{}
			var err error

			if rules != nil {
				doors, err = rules.Eval(*c)
				if err != nil {
					return nil, err
				}
			}

			permission := types.Permissions{
				CardNumber: uint32(*c.Card),
				From:       *c.From,
				To:         *c.To,
				Doors:      doors,
			}

			list = append(list, permission)
		}
	}

	return list, nil
}

func (cc *fdb) add(auth auth.OpAuth, c CardHolder) (*CardHolder, error) {
	oid := catalog.NewCard()

	record := c.clone()
	record.OID = oid
	record.Created = time.Now()

	if auth != nil {
		if err := auth.CanAddCard(record); err != nil {
			return nil, err
		}
	}

	return record, nil
}

func validate(cc fdb) error {
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

func (cc *fdb) scrub() error {
	return nil
}

func unpack(m map[string]interface{}) ([]CardHolder, error) {
	o := struct {
		CardHolders []struct {
			ID     string
			Name   *types.Name
			Card   *types.Card
			From   *types.Date
			To     *types.Date
			Groups map[string]bool
		}
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, err
	}

	cardholders := []CardHolder{}

	for _, r := range o.CardHolders {
		record := CardHolder{
			OID:    catalog.OID(strings.TrimSpace(r.ID)),
			Groups: map[string]bool{},
		}

		record.Name = r.Name.Copy()
		record.Card = r.Card.Copy()
		record.From = r.From.Copy()
		record.To = r.To.Copy()

		for gid, v := range r.Groups {
			record.Groups[gid] = v
		}

		cardholders = append(cardholders, record)
	}

	return cardholders, nil
}

func (d *fdb) log(op string, info interface{}, auth auth.OpAuth) {
	if trail != nil {
		uid := ""
		if auth != nil {
			uid = auth.UID()
		}

		trail.Write(audit.LogEntry{
			UID:       uid,
			Module:    "memdb",
			Operation: op,
			Info:      info,
		})
	}
}
