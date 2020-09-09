package memdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
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
	Groups      []*db.Group      `json:"-"`
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

func NewDB() *fdb {
	groups := []*db.Group{
		&db.Group{ID: "G01", Name: "Teacher"},
		&db.Group{ID: "G02", Name: "Staff"},
		&db.Group{ID: "G03", Name: "Student"},
		&db.Group{ID: "G04", Name: "Gryffindor"},
		&db.Group{ID: "G05", Name: "Hufflepuff"},
		&db.Group{ID: "G06", Name: "Ravenclaw"},
		&db.Group{ID: "G07", Name: "Slytherin"},
		&db.Group{ID: "G08", Name: "Mage"},
		&db.Group{ID: "G09", Name: "Muggle"},
		&db.Group{ID: "G10", Name: "Pet"},
	}

	cardholders := []*db.CardHolder{
		&db.CardHolder{
			ID:         "C01",
			Name:       "Albus Dumbledore",
			CardNumber: 1000101,
			From:       time.Now(),
			To:         time.Now(),
			Groups: []*db.BoolVar{
				&db.BoolVar{ID: "C01G01", Value: false},
				&db.BoolVar{ID: "C01G02", Value: false},
				&db.BoolVar{ID: "C01G03", Value: false},
				&db.BoolVar{ID: "C01G04", Value: true},
				&db.BoolVar{ID: "C01G05", Value: false},
				&db.BoolVar{ID: "C01G06", Value: false},
				&db.BoolVar{ID: "C01G07", Value: false},
				&db.BoolVar{ID: "C01G08", Value: false},
				&db.BoolVar{ID: "C01G09", Value: false},
				&db.BoolVar{ID: "C01G10", Value: false},
			},
		},
		//		&db.CardHolder{ID: "C02", Name: "Tom Riddle", CardNumber: 2000101, From: today(), To: today(), Groups: []*db.BoolVar{}},
		//		&db.CardHolder{ID: "C03", Name: "Harry Potter", CardNumber: 6000101, From: today(), To: today(), Groups: []*db.BoolVar{}},
	}

	return &fdb{
		file: "/usr/local/var/com.github.uhppoted/httpd/memdb/db.json",
		data: data{
			Tables: tables{
				Groups:      groups,
				CardHolders: cardholders,
			},
		},
	}
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
		gid := k

		if value, ok := v.(bool); ok {
			for _, c := range shadow.Tables.CardHolders {
				for _, g := range c.Groups {
					if g.ID == gid {
						g.Value = value
						list.Updated[gid] = g.Value
					}
				}
			}
		}
	}

	if err := save(shadow, d.file); err != nil {
		return list, err
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
