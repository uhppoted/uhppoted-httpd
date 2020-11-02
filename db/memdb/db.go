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

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type fdb struct {
	sync.RWMutex
	file  string
	data  data
	audit audit.Trail
}

type data struct {
	Tables tables `json:"tables"`
}

type tables struct {
	Groups      types.Groups      `json:"groups"`
	CardHolders types.CardHolders `json:"cardholders"`
}

type result struct {
	Updated []interface{} `json:"updated"`
	Deleted []interface{} `json:"deleted"`
}

func (d *data) copy() *data {
	shadow := data{
		Tables: tables{
			Groups:      d.Tables.Groups.Clone(),
			CardHolders: types.CardHolders{},
		},
	}

	for cid, v := range d.Tables.CardHolders {
		shadow.Tables.CardHolders[cid] = v.Clone()
	}

	return &shadow
}

func NewDB(file string, trail audit.Trail) (*fdb, error) {
	f := fdb{
		file: file,
		data: data{
			Tables: tables{
				Groups:      types.Groups{},
				CardHolders: types.CardHolders{},
			},
		},
		audit: trail,
	}

	if err := load(&f.data, f.file); err != nil {
		return nil, err
	}

	return &f, nil
}

func (d *fdb) Groups() types.Groups {
	d.RLock()

	defer d.RUnlock()

	return d.data.Tables.Groups
}

func (d *fdb) CardHolders() (types.CardHolders, error) {
	d.RLock()

	defer d.RUnlock()

	list := types.CardHolders{}

	for cid, record := range d.data.Tables.CardHolders {
		list[cid] = record.Clone()
	}

	return list, nil
}

func (d *fdb) ACL() ([]types.Permissions, error) {
	d.RLock()

	defer d.RUnlock()

	list := []types.Permissions{}

	for _, c := range d.data.Tables.CardHolders {
		if c.Card.IsValid() && c.From.IsValid() && c.To.IsValid() {
			card := uint32(*c.Card)
			from := *c.From
			to := *c.To
			doors := []string{}

			for gid, p := range c.Groups {
				if p {
					if group, ok := d.data.Tables.Groups[gid]; ok {
						doors = append(doors, group.Doors...)
					}
				}
			}

			list = append(list, types.Permissions{
				CardNumber: card,
				From:       from,
				To:         to,
				Doors:      doors,
			})
		}
	}

	return list, nil
}

func (d *fdb) Post(m map[string]interface{}, auth db.IAuth) (interface{}, error) {
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
		if c.ID == "" {
			return nil, &types.HttpdError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("Invalid cardholder ID"),
				Detail: fmt.Errorf("Invalid 'post' request (%w)", fmt.Errorf("Invalid cardholder ID '%v'", c.ID)),
			}
		}

		for _, record := range d.data.Tables.CardHolders {
			if record.ID == c.ID {
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

func (d *fdb) add(shadow *data, ch types.CardHolder, auth db.IAuth) (interface{}, error) {
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

	shadow.Tables.CardHolders[record.ID] = record
	d.log("add", record, auth)

	return record, nil
}

func (d *fdb) update(shadow *data, ch types.CardHolder, auth db.IAuth) (interface{}, error) {
	if record, ok := shadow.Tables.CardHolders[ch.ID]; ok {
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

		for gid, gg := range ch.Groups {
			if _, ok := shadow.Tables.Groups[gid]; ok {
				record.Groups[gid] = gg
			}
		}

		current := d.data.Tables.CardHolders[ch.ID]
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

func (d *fdb) delete(shadow *data, ch types.CardHolder, auth db.IAuth) (interface{}, error) {
	if record, ok := shadow.Tables.CardHolders[ch.ID]; ok {
		if auth != nil {
			if err := auth.CanDeleteCardHolder(record); err != nil {
				return nil, &types.HttpdError{
					Status: http.StatusUnauthorized,
					Err:    fmt.Errorf("Not authorized to delete card holder"),
					Detail: err,
				}
			}
		}

		delete(shadow.Tables.CardHolders, ch.ID)

		d.log("delete", record, auth)

		return record, nil
	}

	return nil, nil
}

func save(d *data, file string) error {
	if err := validate(d); err != nil {
		return err
	}

	if err := clean(d); err != nil {
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
				Detail: fmt.Errorf("record %v: Card holder and card number cannot both be blank", r.ID),
			}
		}

		if r.Card != nil {
			card := uint32(*r.Card)
			if id, ok := cards[card]; ok {
				return &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate card number (%v)", card),
					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", card, id, r.ID),
				}
			}

			cards[card] = r.ID
		}
	}

	return nil
}

func clean(d *data) error {
	for _, r := range d.Tables.CardHolders {
		for gid, _ := range r.Groups {
			if _, ok := d.Tables.Groups[gid]; !ok {
				delete(r.Groups, gid)
			}
		}
	}

	return nil
}

func unpack(m map[string]interface{}) ([]types.CardHolder, error) {
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

	cardholders := []types.CardHolder{}

	for _, r := range o.CardHolders {
		record := types.CardHolder{
			ID:     strings.TrimSpace(r.ID),
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

func (d *fdb) log(op string, info interface{}, auth db.IAuth) {
	if d.audit != nil {
		uid := ""
		if auth != nil {
			uid = auth.UID()
		}

		d.audit.Write(audit.LogEntry{
			UID:       uid,
			Module:    "memdb",
			Operation: op,
			Info:      info,
		})
	}
}
