package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
)

type fdb struct {
	Groups      []db.Group
	CardHolders []db.CardHolder
}

func today() *time.Time {
	d := time.Now()

	return &d
}

func NewDB() *fdb {
	groups := []db.Group{
		db.Group{1, "Teacher"},
		db.Group{2, "Staff"},
		db.Group{3, "Student"},
		db.Group{4, "Gryffindor"},
		db.Group{5, "Hufflepuff"},
		db.Group{6, "Ravenclaw"},
		db.Group{7, "Slytherin"},
		db.Group{8, "Mage"},
		db.Group{9, "Muggle"},
		db.Group{10, "Pet"},
	}

	cardholders := []db.CardHolder{
		db.CardHolder{1, "Albus Dumbledore", 1000101, today(), today(), make([]bool, len(groups))},
		db.CardHolder{2, "Tom Riddle", 2000101, today(), today(), make([]bool, len(groups))},
		db.CardHolder{3, "Harry Potter", 6000101, today(), today(), make([]bool, len(groups))},
	}

	cardholders[0].Groups[3] = true

	return &fdb{
		Groups:      groups,
		CardHolders: cardholders,
	}
}

func (d *fdb) Update(u map[string]interface{}) error {
	if len(u) == 1 {
		return fmt.Errorf("WTF?????")
	}

	re := regexp.MustCompile("G([0-9]+)_([0-9]+)")
	for k, v := range u {
		if match := re.FindStringSubmatch(k); len(match) == 3 {
			cid, _ := strconv.ParseUint(match[1], 10, 32)
			gid, _ := strconv.ParseUint(match[2], 10, 32)

			if value, ok := v.(bool); ok {
				for _, c := range d.CardHolders {
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
