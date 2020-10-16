package memdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	Groups      []*db.Group      `json:"groups"`
	CardHolders []*db.CardHolder `json:"cardholders"`
}

func (d *data) copy() *data {
	shadow := data{
		Tables: tables{
			Groups:      make([]*db.Group, len(d.Tables.Groups)),
			CardHolders: make([]*db.CardHolder, len(d.Tables.CardHolders)),
		},
	}

	for i, v := range d.Tables.Groups {
		shadow.Tables.Groups[i] = v.Copy()
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
				Groups:      []*db.Group{},
				CardHolders: []*db.CardHolder{},
			},
		},
	}

	if err := load(&f.data, f.file); err != nil {
		return nil, err
	}

	return &f, nil
}

func (d *fdb) Groups() []*db.Group {
	d.RLock()

	defer d.RUnlock()

	return d.data.Tables.Groups
}

func (d *fdb) CardHolders() ([]*db.CardHolder, error) {
	d.RLock()

	defer d.RUnlock()

	list := []*db.CardHolder{}

	for _, record := range d.data.Tables.CardHolders {
		name := db.Name{}
		if record.Name != nil {
			name.ID = record.Name.ID
			name.Name = record.Name.Name
		}

		card := db.Card{}
		if record.Card != nil {
			card.ID = record.Card.ID
			card.Number = record.Card.Number
		}

		from := types.Date{}
		if record.From != nil {
			from.ID = record.From.ID
			from.Date = record.From.Date
		}

		to := types.Date{}
		if record.To != nil {
			to.ID = record.To.ID
			to.Date = record.To.Date
		}

		groups := []*db.Permission{}
		for _, g := range d.data.Tables.Groups {
			groups = append(groups, &db.Permission{
				GID:   g.ID,
				Value: false,
			})
		}

		for _, g := range groups {
			for _, gg := range record.Groups {
				if g.GID == gg.GID {
					g.ID = gg.ID
					g.Value = gg.Value
				}
			}
		}

		list = append(list, &db.CardHolder{
			ID:     record.ID,
			Name:   &name,
			Card:   &card,
			From:   &from,
			To:     &to,
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
		ID   string
		Name *struct {
			ID   string
			Name string
		}
		Card *struct {
			ID     string
			Number uint32
		}
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

	if (o.Name == nil || o.Name.Name == "") && (o.Card == nil || o.Card.Number <= 0) {
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

	if o.Name != nil {
		record.Name = &db.Name{
			ID:   o.Name.ID,
			Name: o.Name.Name,
		}
	}

	if o.Card != nil {
		record.Card = &db.Card{
			ID:     o.Card.ID,
			Number: o.Card.Number,
		}
	}

	if o.From != nil {
		record.From = &types.Date{
			ID:   o.From.ID,
			Date: o.From.Date,
		}
	}

	if o.To != nil {
		record.To = &types.Date{
			ID:   o.To.ID,
			Date: o.To.Date,
		}
	}

	for _, g := range d.data.Tables.Groups {
		for gid, gg := range o.Groups {
			if gid == g.ID {
				record.Groups = append(record.Groups, &db.Permission{
					ID:    gg.ID,
					GID:   g.ID,
					Value: gg.Member,
				})
				break
			}
		}
	}

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

func (d *fdb) update(id string, u map[string]interface{}) (interface{}, error) {
	list := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{},
	}

	// update 'shadow' copy

	shadow := d.data.copy()

	for k, v := range u {
		id := k

		for _, c := range shadow.Tables.CardHolders {
			if c.Name.ID == id {
				if value, ok := v.(string); ok {
					c.Name.Name = strings.TrimSpace(value)
					list.Updated[id] = c.Name.Name
					continue
				}

				return nil, &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid card holder name (%v)", v),
					Detail: fmt.Errorf("Error parsing card holder name for card %v - string, got:%v", id, v),
				}
			}

			if c.Card.ID == id {
				if value, ok := v.(uint32); ok {
					c.Card.Number = uint32(value)
					list.Updated[id] = c.Card.Number
					continue
				}

				if _, ok := v.(string); ok {
					value, err := strconv.ParseUint(v.(string), 10, 32)
					if err != nil {
						return nil, &types.HttpdError{
							Status: http.StatusBadRequest,
							Err:    fmt.Errorf("Invalid card number (%v)", v),
							Detail: fmt.Errorf("Error parsing card number %v: %w", v, err),
						}
					}

					c.Card.Number = uint32(value)
					list.Updated[id] = c.Card.Number
					continue
				}

				return nil, &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid card number (%v)", v),
					Detail: fmt.Errorf("Error parsing card number for card %v - expected:uint32/string, got:%v", id, v),
				}
			}

			if c.From.ID == id {
				if _, ok := v.(string); ok {
					value, err := time.Parse("2006-01-02", v.(string))
					if err != nil {
						return nil, &types.HttpdError{
							Status: http.StatusBadRequest,
							Err:    fmt.Errorf("Invalid 'from' date (%v)", v),
							Detail: fmt.Errorf("Error parsing 'from' date %v: %w", v, err),
						}
					}

					c.From.Date = value
					list.Updated[id] = c.From.Format("2006-01-02")
					continue
				}

				return nil, &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid 'from' date (%v)", v),
					Detail: fmt.Errorf("Error parsing 'from' date %v for card - expected:YYYY-MM-DD, got:%v", id, v),
				}
			}

			if c.To.ID == id {
				if _, ok := v.(string); ok {
					value, err := time.Parse("2006-01-02", v.(string))
					if err != nil {
						return nil, &types.HttpdError{
							Status: http.StatusBadRequest,
							Err:    fmt.Errorf("Invalid 'to' date (%v)", v),
							Detail: fmt.Errorf("Error parsing 'to' date %v: %w", v, err),
						}
					}

					c.To.Date = value
					list.Updated[id] = c.To.Format("2006-01-02")
					continue
				}

				return nil, &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Invalid 'to' date (%v)", v),
					Detail: fmt.Errorf("Error parsing 'to' date %v for card - expected:YYYY-MM-DD, got:%v", id, v),
				}
			}

			for _, g := range c.Groups {
				if g.ID == id {
					if value, ok := v.(bool); ok {
						g.Value = value
						list.Updated[id] = g.Value
						continue
					}

					if value, ok := v.(string); ok {
						if b, err := strconv.ParseBool(value); err == nil {
							g.Value = b
							list.Updated[id] = g.Value
							continue
						}
					}

					return nil, &types.HttpdError{
						Status: http.StatusBadRequest,
						Err:    fmt.Errorf("Invalid group value (%v)", v),
						Detail: fmt.Errorf("Error parsing group value for group-id %v - expected:bool, got:%v", id, v),
					}
				}
			}
		}
	}

	// check integrity of updated data

	N := len(shadow.Tables.CardHolders)
	for i := 0; i < N; i++ {
		p := shadow.Tables.CardHolders[i]
		for j := i + 1; j < N; j++ {
			q := shadow.Tables.CardHolders[j]
			if p.Card.Number == q.Card.Number {
				return nil, &types.HttpdError{
					Status: http.StatusBadRequest,
					Err:    fmt.Errorf("Duplicate card number (%v)", p.Card.Number),
					Detail: fmt.Errorf("card %v: duplicate entry in records %v and %v", p.Card.Number, p.ID, q.ID),
				}
			}
		}
	}

	if err := save(shadow, d.file); err != nil {
		return nil, err
	}

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
