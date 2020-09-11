package system

import (
	"fmt"
	"log"
	"strings"

	api "github.com/uhppoted/uhppoted-api/acl"
	acl "github.com/uhppoted/uhppoted-httpd/acl"
)

type System interface {
	Update(permissions acl.ACL)
}

var local = Local{}

func Update(permissions acl.ACL) {
	local.Update(permissions)
}

func permissionsToTable(p acl.ACL) (*api.Table, error) {
	header := []string{"Card Number", "From", "To"}
	records := [][]string{}
	index := map[string]int{}

	list := p.Permissions()

	for _, q := range list {
		fmt.Printf(">> PERMISSION %v", q)
		if q.From == nil {
			return nil, fmt.Errorf("Card %v: missing 'start-date'", q.CardNumber)
		}

		if q.To == nil {
			return nil, fmt.Errorf("Card %v: missing 'end-date'", q.CardNumber)
		}

		for _, door := range q.Doors {
			d := clean(door)
			if _, ok := index[d]; !ok {
				index[d] = 3 + len(index)
				header = append(header, door)
			}
		}
	}

	// for _, r := range p {
	//     record := make([]string, len(header))
	//     record[0] = fmt.Sprintf("%v", r.CardNumber)
	//     record[1] = fmt.Sprintf("%s", r.From)
	//     record[2] = fmt.Sprintf("%s", r.To)
	//     for i := 3; i < len(record); i++ {
	//         record[i] = "N"
	//     }

	//     for _, door := range r.Doors {
	//         d := clean(door)
	//         if ix, ok := index[d]; !ok {
	//             return nil, fmt.Errorf("Card %v: unindexed door '%s'", r.CardNumber, door)
	//         } else {
	//             record[ix] = "Y"
	//         }
	//     }

	//     records = append(records, record)
	// }

	table := api.Table{
		Header:  header,
		Records: records,
	}

	return &table, nil
}

func clean(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
