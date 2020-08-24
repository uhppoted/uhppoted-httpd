package httpd

import (
	"time"
)

type db struct {
	Groups      []*Group
	CardHolders []*CardHolder
}

type Group struct {
	ID   uint32
	Name string
}

type CardHolder struct {
	ID         uint32
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
	groups := []*Group{
		&Group{1, "Teacher"},
		&Group{2, "Staff"},
		&Group{3, "Student"},
		&Group{4, "Gryffindor"},
		&Group{5, "Hufflepuff"},
		&Group{6, "Ravenclaw"},
		&Group{7, "Slytherin"},
		&Group{8, "Mage"},
		&Group{9, "Muggle"},
		&Group{10, "Pet"},
	}

	cardholders := []*CardHolder{
		&CardHolder{1, "Albus Dumbledore", 1000101, today(), today(), make([]bool, len(groups))},
		&CardHolder{2, "Tom Riddle", 2000101, today(), today(), make([]bool, len(groups))},
		&CardHolder{3, "Harry Potter", 6000101, today(), today(), make([]bool, len(groups))},
	}

	cardholders[0].Groups[3] = true

	return &db{
		Groups:      groups,
		CardHolders: cardholders,
	}
}
