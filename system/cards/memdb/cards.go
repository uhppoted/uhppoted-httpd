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
	Tables tables `json:"tables"`
}

type tables struct {
	CardHolders cards.CardHolders `json:"cardholders"`
}

type result struct {
	Updated []interface{} `json:"updated"`
	Deleted []interface{} `json:"deleted"`
}

const GroupName = catalog.GroupName

var trail audit.Trail

func (d *data) copy() *data {
	shadow := data{
		Tables: tables{
			CardHolders: cards.CardHolders{},
		},
	}

	for cid, v := range d.Tables.CardHolders {
		shadow.Tables.CardHolders[cid] = v.Clone()
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
			Tables: tables{
				//Groups:      types.Groups{},
				CardHolders: cards.CardHolders{},
			},
		},
		rules: rules,
	}

	if err := load(&f.data, f.file); err != nil {
		return nil, err
	}

	created := time.Time{}

	for _, c := range f.data.Tables.CardHolders {
		created = created.Add(1 * time.Minute)
		c.Created = created
	}

	return &f, nil
}

func (d *fdb) Clone() cards.Cards {
	shadow := data{
		Tables: tables{
			CardHolders: cards.CardHolders{},
		},
	}

	for cid, v := range d.data.Tables.CardHolders {
		shadow.Tables.CardHolders[cid] = v.Clone()
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

	for k, c := range cc.data.Tables.CardHolders {
		if c.OID.Contains(oid) {
			objects, err := c.Set(auth, oid, value)
			if err == nil {
				cc.data.Tables.CardHolders[k] = c
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	//	if oid == "<new>" {
	//		if d, err := dd.add(auth, Door{}); err != nil {
	//			return nil, err
	//		} else if d == nil {
	//			return nil, fmt.Errorf("Failed to add 'new' door")
	//		} else {
	//			d.log(auth, "add", d.OID, "door", "", "")
	//			dd.Doors[d.OID] = *d
	//			objects = append(objects, object{
	//				OID:   d.OID,
	//				Value: "new",
	//			})
	//		}
	//	}

	return objects, nil
}

func (d *fdb) CardHolders() cards.CardHolders {
	d.RLock()

	defer d.RUnlock()

	list := cards.CardHolders{}

	for cid, record := range d.data.Tables.CardHolders {
		list[cid] = record.Clone()
	}

	return list
}

func (d *fdb) Print() {
	if b, err := json.MarshalIndent(d.data.Tables.CardHolders, "", "  "); err == nil {
		fmt.Printf("----------------- CARDS\n%s\n", string(b))
	}
}

func (d *fdb) AsObjects() []interface{} {
	objects := []interface{}{}

	d.RLock()

	defer d.RUnlock()

	for _, record := range d.data.Tables.CardHolders {
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

	for _, c := range d.data.Tables.CardHolders {
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

func (d *fdb) Post(m map[string]interface{}, auth auth.OpAuth) (interface{}, error) {
	d.Lock()

	defer d.Unlock()

	// add/update ?

	cardholders, err := unpack(m)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid request"),
			Detail: fmt.Errorf("Error unpacking 'post' request (%w)", err),
		}
	}

	list := result{}
	shadow := d.data.copy()

loop:
	for _, c := range cardholders {
		if c.OID == "" {
			return nil, &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid cardholder ID"),
				Detail: fmt.Errorf("Invalid 'post' request (%w)", fmt.Errorf("Invalid cardholder ID '%v'", c.OID)),
			}
		}

		for _, record := range d.data.Tables.CardHolders {
			if record.OID == c.OID {
				if c.Name != nil && *c.Name == "" && c.Card != nil && *c.Card == 0 {
					if r, err := d.delete(shadow, c, auth); err != nil {
						return nil, err
					} else if r != nil {
						list.Deleted = append(list.Deleted, r)
						continue loop
					}
				}

				if r, err := d.update(shadow, c, auth); err != nil {
					return nil, err
				} else if r != nil {
					list.Updated = append(list.Updated, r)
				}

				continue loop
			}
		}

		if r, err := d.add(shadow, c, auth); err != nil {
			return nil, err
		} else if r != nil {
			list.Updated = append(list.Updated, r)
		}
	}

	if err := save(shadow, d.file); err != nil {
		return nil, err
	}

	d.data = *shadow

	return list, nil
}

func (d *fdb) add(shadow *data, ch cards.CardHolder, auth auth.OpAuth) (interface{}, error) {
	record := ch.Clone()

	if auth != nil {
		if err := auth.CanAddCardHolder(record); err != nil {
			return nil, &types.HttpdError{
				Status: http.StatusUnauthorized,
				Err:    fmt.Errorf("Not authorized to add card holder"),
				Detail: err,
			}
		}
	}

	shadow.Tables.CardHolders[record.OID] = record
	d.log("add", record, auth)

	return record, nil
}

func (d *fdb) update(shadow *data, ch cards.CardHolder, auth auth.OpAuth) (interface{}, error) {
	if record, ok := shadow.Tables.CardHolders[ch.OID]; ok {
		if ch.Name != nil {
			record.Name = ch.Name
		}

		if ch.Card != nil {
			record.Card = ch.Card
		}

		if ch.From != nil {
			record.From = ch.From
		}

		if ch.To != nil {
			record.To = ch.To
		}

		current := d.data.Tables.CardHolders[ch.OID]
		if auth != nil {
			if err := auth.CanUpdateCardHolder(current, record); err != nil {
				return nil, &types.HttpdError{
					Status: http.StatusUnauthorized,
					Err:    fmt.Errorf("Not authorized to update card holder"),
					Detail: err,
				}
			}
		}

		d.log("update", map[string]interface{}{"original": current, "updated": record}, auth)

		return record, nil
	}

	return nil, nil
}

func (d *fdb) delete(shadow *data, ch cards.CardHolder, auth auth.OpAuth) (interface{}, error) {
	if record, ok := shadow.Tables.CardHolders[ch.OID]; ok {
		if auth != nil {
			if err := auth.CanDeleteCardHolder(record); err != nil {
				return nil, &types.HttpdError{
					Status: http.StatusUnauthorized,
					Err:    fmt.Errorf("Not authorized to delete card holder"),
					Detail: err,
				}
			}
		}

		delete(shadow.Tables.CardHolders, ch.OID)

		d.log("delete", record, auth)

		return record, nil
	}

	return nil, nil
}

func save(d *data, file string) error {
	if err := validate(d); err != nil {
		return err
	}

	if err := scrub(d); err != nil {
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

func validate(d *data) error {
	cards := map[uint32]string{}

	for _, r := range d.Tables.CardHolders {
		if !r.Name.IsValid() && !r.Card.IsValid() {
			return &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Name and card number cannot both be empty"),
				Detail: fmt.Errorf("record %v: Card holder and card number cannot both be blank", r.OID),
			}
		}

		if r.Card != nil {
			card := uint32(*r.Card)
			if id, ok := cards[card]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate card number (%v)", card),
					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", card, id, r.OID),
				}
			}

			cards[card] = string(r.OID)
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
