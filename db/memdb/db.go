package memdb

import (
	"encoding/json"
	"fmt"
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
	Groups      []db.Group       `json:"groups"`
	CardHolders []*db.CardHolder `json:"cardholders"`
}

func today() *time.Time {
	d := time.Now()

	return &d
}

func NewDB() *fdb {
	groups := []db.Group{
		db.Group{ID: "G01", Name: "Teacher"},
		db.Group{ID: "G02", Name: "Staff"},
		db.Group{ID: "G03", Name: "Student"},
		db.Group{ID: "G04", Name: "Gryffindor"},
		db.Group{ID: "G05", Name: "Hufflepuff"},
		db.Group{ID: "G06", Name: "Ravenclaw"},
		db.Group{ID: "G07", Name: "Slytherin"},
		db.Group{ID: "G08", Name: "Mage"},
		db.Group{ID: "G09", Name: "Muggle"},
		db.Group{ID: "G10", Name: "Pet"},
	}

	cardholders := []*db.CardHolder{
		&db.CardHolder{ID: "C01", Name: "Albus Dumbledore", CardNumber: 1000101, From: today(), To: today(), Groups: []*db.BoolVar{}},
		&db.CardHolder{ID: "C02", Name: "Tom Riddle", CardNumber: 2000101, From: today(), To: today(), Groups: []*db.BoolVar{}},
		&db.CardHolder{ID: "C03", Name: "Harry Potter", CardNumber: 6000101, From: today(), To: today(), Groups: []*db.BoolVar{}},
	}

	for _, c := range cardholders {
		for _, g := range groups {
			c.Groups = append(c.Groups, &db.BoolVar{
				ID:    fmt.Sprintf("%v%v", c.ID, g.ID),
				Value: false,
			})
		}
	}

	cardholders[0].Groups[3].Value = true

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

func (d *fdb) Groups() []db.Group {
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

	updated := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{},
	}

	for k, v := range u {
		gid := k

		if value, ok := v.(bool); ok {
			for _, c := range d.data.Tables.CardHolders {
				for _, g := range c.Groups {
					if g.ID == gid {
						g.Value = value
						updated.Updated[gid] = g.Value
					}
				}
			}
		}
	}

	if err := d.save(); err != nil {
		return updated, err
	}

	return updated, nil
}

func (d *fdb) save() error {
	b, err := json.Marshal(d.data)
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

	if err := os.MkdirAll(filepath.Dir(d.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), d.file)
}
