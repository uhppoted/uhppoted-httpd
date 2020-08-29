package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
)

type fdb struct {
	sync.RWMutex
	groups      []db.Group
	cardHolders []db.CardHolder
}

func today() *time.Time {
	d := time.Now()

	return &d
}

func NewDB() *fdb {
	groups := []db.Group{
		db.Group{ID: 1, Name: "Teacher"},
		db.Group{ID: 2, Name: "Staff"},
		db.Group{ID: 3, Name: "Student"},
		db.Group{ID: 4, Name: "Gryffindor"},
		db.Group{ID: 5, Name: "Hufflepuff"},
		db.Group{ID: 6, Name: "Ravenclaw"},
		db.Group{ID: 7, Name: "Slytherin"},
		db.Group{ID: 8, Name: "Mage"},
		db.Group{ID: 9, Name: "Muggle"},
		db.Group{ID: 10, Name: "Pet"},
	}

	cardholders := []db.CardHolder{
		db.CardHolder{ID: 1, Name: "Albus Dumbledore", CardNumber: 1000101, From: today(), To: today(), Groups: make([]bool, len(groups))},
		db.CardHolder{ID: 2, Name: "Tom Riddle", CardNumber: 2000101, From: today(), To: today(), Groups: make([]bool, len(groups))},
		db.CardHolder{ID: 3, Name: "Harry Potter", CardNumber: 6000101, From: today(), To: today(), Groups: make([]bool, len(groups))},
	}

	cardholders[0].Groups[3] = true

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

func (d *fdb) CardHolders() []db.CardHolder {
	d.RLock()

	defer d.RUnlock()

	return d.cardHolders
}

func (d *fdb) Update(u map[string]interface{}) error {
	if len(u) == 1 {
		return fmt.Errorf("WTF?????")
	}

	d.Lock()

	defer d.Unlock()

	re := regexp.MustCompile("G([0-9]+)_([0-9]+)")
	for k, v := range u {
		if match := re.FindStringSubmatch(k); len(match) == 3 {
			cid, _ := strconv.ParseUint(match[1], 10, 32)
			gid, _ := strconv.ParseUint(match[2], 10, 32)

			if value, ok := v.(bool); ok {
				for _, c := range d.cardHolders {
					if c.ID == uint32(cid) {
						for ix, _ := range c.Groups {
							if uint32(ix) == uint32(gid) {
								c.Groups[ix] = value
							}
						}
					}
				}
			}
		}
	}

	return nil
}
