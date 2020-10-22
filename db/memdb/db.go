package memdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	//	"path/filepath"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type fdb struct {
	sync.RWMutex
	file string
	data data
}

type data struct {
	Tables tables `json:"tables"`
}

type tables struct {
	Groups      types.Groups      `json:"groups"`
	CardHolders types.CardHolders `json:"cardholders"`
}

type result struct {
	Added   []interface{} `json:"added"`
	Updated []interface{} `json:"updated"`
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

func NewDB(file string) (*fdb, error) {
	f := fdb{
		file: file,
		data: data{
			Tables: tables{
				Groups:      types.Groups{},
				CardHolders: types.CardHolders{},
			},
		},
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

func (d *fdb) Post(id string, u map[string]interface{}) (interface{}, error) {
	d.Lock()

	defer d.Unlock()

	fmt.Printf("post: %v\n", u)

	// add/update ?

	var record *types.CardHolder

	for _, c := range d.data.Tables.CardHolders {
		if c.ID == id {
			record = c
			break
		}
	}

	if record != nil {
		return d.update(id, u)
	}

	return d.add(id, u)
}

func (d *fdb) add(id string, m map[string]interface{}) (interface{}, error) {
	record, err := unpack(id, m)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid 'add' request"),
			Detail: fmt.Errorf("Error unpacking 'add' request (%w)", err),
		}
	}

	// ... append to DB

	list := result{}
	shadow := d.data.copy()

	shadow.Tables.CardHolders[id] = record

	if err := save(shadow, d.file); err != nil {
		return nil, err
	}

	d.data = *shadow

	list.Added = append(list.Added, record)

	return list, nil
}

func (d *fdb) update(id string, m map[string]interface{}) (interface{}, error) {
	r, err := unpack(id, m)
	if err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid 'update' request"),
			Detail: fmt.Errorf("Error unpacking 'update' request (%w)", err),
		}
	}

	// update 'shadow' copy

	var record *types.CardHolder

	list := result{}
	shadow := d.data.copy()

	for _, c := range shadow.Tables.CardHolders {
		if c.ID == id {
			record = c
			break
		}
	}

	if record != nil {
		if r.Name != nil {
			record.Name = r.Name
		}

		if r.Card != nil {
			record.Card = r.Card
		}

		if r.From != nil {
			record.From = r.From
		}

		if r.To != nil {
			record.To = r.To
		}

		for gid, gg := range r.Groups {
			if _, ok := shadow.Tables.Groups[gid]; ok {
				record.Groups[gid] = gg
			}
		}

		if err := save(shadow, d.file); err != nil {
			return nil, err
		}

		d.data = *shadow

		list.Updated = append(list.Updated, record)
	}

	return list, nil
}

func save(d *data, file string) error {
	if err := validate(d); err != nil {
		return err
	}

	if err := clean(d); err != nil {
		return err
	}

	return nil
	//	if file == "" {
	//		return nil
	//	}
	//
	//	b, err := json.MarshalIndent(d, "", "  ")
	//	if err != nil {
	//		return err
	//	}
	//
	//	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted-*.db")
	//	if err != nil {
	//		return err
	//	}
	//
	//	defer os.Remove(tmp.Name())
	//
	//	if _, err := tmp.Write(b); err != nil {
	//		return err
	//	}
	//
	//	if err := tmp.Close(); err != nil {
	//		return err
	//	}
	//
	//	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
	//		return err
	//	}
	//
	//	return os.Rename(tmp.Name(), file)
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

func unpack(id string, m map[string]interface{}) (*types.CardHolder, error) {
	o := struct {
		Name   *types.Name
		Card   *types.Card
		From   *types.Date
		To     *types.Date
		Groups map[string]bool
	}{}

	blob, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(blob, &o); err != nil {
		return nil, err
	}

	record := types.CardHolder{
		ID:     id,
		Groups: map[string]bool{},
	}

	record.Name = o.Name.Copy()
	record.Card = o.Card.Copy()
	record.From = o.From.Copy()
	record.To = o.To.Copy()

	for gid, v := range o.Groups {
		record.Groups[gid] = v
	}

	return &record, nil
}
