package memdb

import (
	"fmt"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
)

type fdb struct {
	sync.RWMutex
	groups      []db.Group
	cardHolders []*db.CardHolder
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
		groups:      groups,
		cardHolders: cardholders,
	}
}

func (d *fdb) Groups() []db.Group {
	d.RLock()

	defer d.RUnlock()

	return d.groups
}

func (d *fdb) CardHolders() []*db.CardHolder {
	d.RLock()

	defer d.RUnlock()

	return d.cardHolders
}

func (d *fdb) Update(u map[string]interface{}) (map[string]interface{}, error) {
	d.Lock()

	defer d.Unlock()

	updated := map[string]interface{}{}

	for k, v := range u {
		gid := k

		if value, ok := v.(bool); ok {
			for _, c := range d.cardHolders {
				for _, g := range c.Groups {
					if g.ID == gid {
						g.Value = value
						updated[gid] = g.Value
					}
				}
			}
		}
	}

	return updated, nil
}
