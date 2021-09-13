package memdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type fdb struct {
	sync.RWMutex
	file  string
	data  data
	rules cards.IRules
}

type data struct {
	CardHolders cards.CardHolders `json:"cardholders"`
}

type result struct {
	Updated []interface{} `json:"updated"`
	Deleted []interface{} `json:"deleted"`
}

type object catalog.Object

const GroupName = catalog.GroupName

var trail audit.Trail

func (d *data) copy() *data {
	shadow := data{
		CardHolders: cards.CardHolders{},
	}

	for cid, v := range d.CardHolders {
		shadow.CardHolders[cid] = v.Clone()
	}

	return &shadow
}

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func NewDB(file string, rules cards.IRules) (*fdb, error) {
	f := fdb{
		file: file,
		data: data{
			CardHolders: cards.CardHolders{},
		},
		rules: rules,
	}

	if err := load(&f.data, f.file); err != nil {
		return nil, err
	}

	created := time.Time{}

	for _, c := range f.data.CardHolders {
		created = created.Add(1 * time.Minute)
		c.Created = created
	}

	for _, c := range f.data.CardHolders {
		catalog.PutCard(c.OID)
	}

	return &f, nil
}

func (d *fdb) Clone() cards.Cards {
	shadow := data{
		CardHolders: cards.CardHolders{},
	}

	for cid, v := range d.data.CardHolders {
		shadow.CardHolders[cid] = v.Clone()
	}

	return &fdb{
		file:  d.file,
		data:  shadow,
		rules: d.rules,
	}
}

func (cc *fdb) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if cc == nil {
		return nil, nil
	}

	for k, c := range cc.data.CardHolders {
		if c.OID.Contains(oid) {
			objects, err := c.Set(auth, oid, value)
			if err == nil {
				cc.data.CardHolders[k] = c
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	if oid == "<new>" {
		if c, err := cc.add(auth, cards.CardHolder{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' card")
		} else {
			c.Log(auth, "add", c.OID, "card", "", "")
			cc.data.CardHolders[c.OID] = c
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
		return validate(cc.data)
	}

	return nil
}

func (d *fdb) CardHolders() cards.CardHolders {
	d.RLock()

	defer d.RUnlock()

	list := cards.CardHolders{}

	for cid, record := range d.data.CardHolders {
		list[cid] = record.Clone()
	}

	return list
}

func (d *fdb) Print() {
	if b, err := json.MarshalIndent(d.data.CardHolders, "", "  "); err == nil {
		fmt.Printf("----------------- CARDS\n%s\n", string(b))
	}
}

func (d *fdb) AsObjects() []interface{} {
	objects := []interface{}{}

	d.RLock()

	defer d.RUnlock()

	for _, record := range d.data.CardHolders {
		if record.IsValid() || record.IsDeleted() {
			if l := record.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (d *fdb) ACL() ([]types.Permissions, error) {
	d.RLock()

	defer d.RUnlock()

	list := []types.Permissions{}

	for _, c := range d.data.CardHolders {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			var doors = []string{}
			var err error

			if d.rules != nil {
				doors, err = d.rules.Eval(*c)
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

func (cc *fdb) add(auth auth.OpAuth, c cards.CardHolder) (*cards.CardHolder, error) {
	oid := catalog.NewCard()

	record := c.Clone()
	record.OID = oid
	record.Created = time.Now()

	if auth != nil {
		if err := auth.CanAddCard(record); err != nil {
			return nil, err
		}
	}

	return record, nil
}

func save(d data, file string) error {
	if err := validate(d); err != nil {
		return err
	}

	if err := scrub(&d); err != nil {
		return err
	}

	if file == "" {
		return nil
	}

	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted-*.db")
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

	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), file)
}

func load(data interface{}, file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return json.Unmarshal(b, data)
}

func validate(d data) error {
	cards := map[uint32]string{}

	for _, c := range d.CardHolders {
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

func scrub(d *data) error {
	return nil
}

func unpack(m map[string]interface{}) ([]cards.CardHolder, error) {
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

	cardholders := []cards.CardHolder{}

	for _, r := range o.CardHolders {
		record := cards.CardHolder{
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
