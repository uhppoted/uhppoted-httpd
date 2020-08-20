package httpd

import (
	"time"
)

type db struct {
	Groups      []string
	CardHolders []*CardHolder
}

type CardHolder struct {
	Name       string
	CardNumber uint32
	From       *time.Time
	To         *time.Time
	Groups     []bool
}

func today() *time.Time {
	d := time.Now()

	return &d
}

func NewDB() *db {
	groups := []string{"Teacher", "Staff", "Student", "Gryffindor", "Hufflepuff", "Ravenclaw", "Slytherin", "Mage", "Muggle", "Pet"}
	return &db{
		Groups: []string{"Teacher", "Staff", "Student", "Gryffindor", "Hufflepuff", "Ravenclaw", "Slytherin", "Mage", "Muggle", "Pet"},
		CardHolders: []*CardHolder{
			&CardHolder{"Albus Dumbledore", 1000101, today(), today(), make([]bool, len(groups))},
			&CardHolder{"Tom Riddle", 2000101, today(), today(), make([]bool, len(groups))},
			&CardHolder{"Harry Potter", 600010, today(), today(), make([]bool, len(groups))},
		},
	}
}
