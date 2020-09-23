package memdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

	return d.data.Tables.CardHolders, nil
}

func (d *fdb) ACL() ([]types.Permissions, error) {
	d.RLock()

	defer d.RUnlock()

	list := []types.Permissions{}

	for _, c := range d.data.Tables.CardHolders {
		doors := []string{}

		for _, p := range c.Groups {
			if p.Value {
				for _, group := range d.data.Tables.Groups {
					if p.GID == group.ID {
						doors = append(doors, group.Doors...)
					}
				}
			}
		}

		list = append(list, types.Permissions{
			CardNumber: c.Card.Number,
			From:       c.From,
			To:         c.To,
			Doors:      doors,
		})
	}

	return list, nil
}

func (d *fdb) Update(u map[string]interface{}) (interface{}, error) {
	d.Lock()

	defer d.Unlock()

	list := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{},
	}

	shadow := d.data.copy()

	for k, v := range u {
		id := k

		for _, c := range shadow.Tables.CardHolders {
			if c.Card.ID == id {
				if _, ok := v.(string); ok {
					if value, err := strconv.ParseUint(v.(string), 10, 32); err != nil {
						return nil, err
					} else {
						c.Card.Number = uint32(value)
						list.Updated[id] = c.Card.Number
					}
				}
			}

			for _, g := range c.Groups {
				if g.ID == id {
					if value, ok := v.(bool); !ok {
						return nil, fmt.Errorf("Invalid value for %v - expected bool, got %v", id, v)
					} else {
						g.Value = value
						list.Updated[id] = g.Value
					}
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
