package memdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/db"
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
	Groups      types.Groups     `json:"groups"`
	CardHolders []*db.CardHolder `json:"cardholders"`
}

func (d *data) copy() *data {
	shadow := data{
		Tables: tables{
			Groups:      d.Tables.Groups.Copy(),
			CardHolders: make([]*db.CardHolder, len(d.Tables.CardHolders)),
		},
	}

	for i, v := range d.Tables.CardHolders {
		shadow.Tables.CardHolders[i] = v.Copy()
	}

	return &shadow
}

func NewDB(file string) (*fdb, error) {
	f := fdb{
		file: file,
		data: data{
			Tables: tables{
				Groups:      types.Groups{},
				CardHolders: []*db.CardHolder{},
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

func (d *fdb) CardHolders() ([]*db.CardHolder, error) {
	d.RLock()

	defer d.RUnlock()

	list := []*db.CardHolder{}

	for _, record := range d.data.Tables.CardHolders {
		name := record.Name.Copy()
		card := record.Card.Copy()
		from := record.From.Copy()
		to := record.To.Copy()

		groups := []*db.Permission{}
		//		for _, g := range d.data.Tables.Groups {
		//			groups = append(groups, &db.Permission{
		//				GID:   g.ID,
		//				Value: false,
		//			})
		//		}
		//
		//		for _, g := range groups {
		//			for _, gg := range record.Groups {
		//				if g.GID == gg.GID {
		//					g[GID] = gg.ID
		//					g.Value = gg.Value
		//				}
		//			}
		//		}

		list = append(list, &db.CardHolder{
			ID:     record.ID,
			Name:   name,
			Card:   card,
			From:   from,
			To:     to,
			Groups: groups,
		})
	}

	return list, nil
}

func (d *fdb) ACL() ([]types.Permissions, error) {
	d.RLock()

	defer d.RUnlock()

	list := []types.Permissions{}

	//	for _, c := range d.data.Tables.CardHolders {
	//		doors := []string{}
	//
	//		for _, p := range c.Groups {
	//			if p.Value {
	//				for _, group := range d.data.Tables.Groups {
	//					if p.GID == group.ID {
	//						doors = append(doors, group.Doors...)
	//					}
	//				}
	//			}
	//		}
	//
	//		list = append(list, types.Permissions{
	//			CardNumber: c.Card.Number,
	//			From:       c.From,
	//			To:         c.To,
	//			Doors:      doors,
	//		})
	//	}

	return list, nil
}

func (d *fdb) Post(id string, u map[string]interface{}) (interface{}, error) {
	d.Lock()

	defer d.Unlock()

	// add/update ?

	var record *db.CardHolder

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
	o := struct {
		ID     string
		Name   *types.Name
		Card   *types.Card
		From   *types.Date
		To     *types.Date
		Groups map[string]struct {
			ID     string
			Member bool
		}
	}{}

	if err := unpack(m, &o); err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid 'add' request"),
			Detail: fmt.Errorf("Error unpacking 'add' request (%w)", err),
		}
	}

	if !o.Name.IsValid() && !o.Card.IsValid() {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Name and card number cannot both be empty"),
			Detail: fmt.Errorf("Card holder and card number cannot both be blank"),
		}
	}

	record := db.CardHolder{
		ID:     id,
		Groups: []*db.Permission{},
	}

	record.Name = o.Name.Copy()
	record.Card = o.Card.Copy()
	record.From = o.From.Copy()
	record.To = o.To.Copy()

	//	for _, g := range d.data.Tables.Groups {
	//		for gid, gg := range o.Groups {
	//			if gid == g.ID {
	//				record.Groups = append(record.Groups, &db.Permission{
	//					ID:    gg.ID,
	//					GID:   g.ID,
	//					Value: gg.Member,
	//				})
	//				break
	//			}
	//		}
	//	}

	// ... append to DB

	list := struct {
		Added map[string]interface{} `json:"added"`
	}{
		Added: map[string]interface{}{},
	}

	shadow := d.data.copy()

	shadow.Tables.CardHolders = append(shadow.Tables.CardHolders, &record)

	// if err := save(shadow, d.file); err != nil {
	// 	return nil, err
	// }

	d.data = *shadow

	return list, nil
}

func (d *fdb) update(id string, m map[string]interface{}) (interface{}, error) {
	o := struct {
		ID     string
		Name   *types.Name
		Card   *types.Card
		From   *types.Date
		To     *types.Date
		Groups map[string]struct {
			ID     string
			Member bool
		}
	}{}

	if err := unpack(m, &o); err != nil {
		return nil, &types.HttpdError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("Invalid 'add' request"),
			Detail: fmt.Errorf("Error unpacking 'add' request (%w)", err),
		}
	}

	// update 'shadow' copy

	var record *db.CardHolder

	list := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{},
	}

	shadow := d.data.copy()

	for _, c := range shadow.Tables.CardHolders {
		if c.ID == id {
			record = c
			break
		}
	}

	fmt.Printf("GOTCHA: %+v\n", record)
	if record != nil {
		if o.Name != nil {
			record.Name = o.Name
		}
		if o.Card != nil {
			record.Card = o.Card
		}
		if o.From != nil {
			record.From = o.From
		}
		if o.To != nil {
			record.To = o.To
		}

		//for gid, gg := range o.Groups {
		//	for _, g := range shadow.Tables.Groups {
		//		if gid == g.ID {
		//			//					record.Groups = append(record.Groups, &db.Permission{
		//			//						ID:    gg.ID,
		//			//						GID:   g.ID,
		//			//						Value: gg.Member,
		//			//					})
		//			break
		//		}
		//	}
		//}
	}
	//
	//	// check integrity of updated data
	//
	//	N := len(shadow.Tables.CardHolders)
	//	for i := 0; i < N; i++ {
	//		p := shadow.Tables.CardHolders[i]
	//		for j := i + 1; j < N; j++ {
	//			q := shadow.Tables.CardHolders[j]
	//			if p.Card != nil && q.Card != nil && p.Card == q.Card {
	//				return nil, &types.HttpdError{
	//					Status: http.StatusBadRequest,
	//					Err:    fmt.Errorf("Duplicate card number (%v)", p.Card),
	//					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", p.Card, p.ID, q.ID),
	//				}
	//			}
	//		}
	//	}
	//
	//	if err := save(shadow, d.file); err != nil {
	//		return nil, err
	//	}

	d.data = *shadow

	return list, nil
}

func save(data interface{}, file string) error {
	if file == "" {
		return nil
	}

	b, err := json.MarshalIndent(data, "", "  ")
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

func unpack(m map[string]interface{}, o interface{}) error {
	blob, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(blob, o)
}
