package system

import (
	"fmt"
	"regexp"
	"strings"
)

type Resolver struct {
}

func (r Resolver) Get(query string) []interface{} {
	switch {
	case strings.HasPrefix(query, "controller.OID for door.OID"):
		return r.lookupControllerForDoor(query)

	case strings.HasPrefix(query, "controller.Created for door.OID"):
		return r.lookupControllerForDoor(query)

	case strings.HasPrefix(query, "controller.Name for door.OID"):
		return r.lookupControllerForDoor(query)

	case strings.HasPrefix(query, "controller.ID for door.OID"):
		return r.lookupControllerForDoor(query)

	case strings.HasPrefix(query, "controller.Door for door.OID"):
		return r.lookupControllerForDoor(query)

	case strings.HasPrefix(query, "controller.Door.Delay for door.OID"):
		return r.lookupControllerForDoor(query)
	}

	return nil
}

func (r Resolver) lookupControllerForDoor(query string) []interface{} {
	re := regexp.MustCompile(`controller.(OID|Created|Name|ID|Door|Door\.Delay) for door.OID\[(.*?)\]`)

	match := re.FindStringSubmatch(query)
	if match == nil || len(match) < 3 {
		return nil
	}

	field := match[1]
	oid := match[2]
	resultset := []interface{}{}

	for _, c := range sys.controllers.Controllers {
		for k, d := range c.Doors {
			if d == oid {
				switch field {
				case "OID":
					resultset = append(resultset, c.Get("OID"))
				case "Created":
					resultset = append(resultset, c.Get("created"))
				case "Name":
					resultset = append(resultset, c.Get("name"))
				case "ID":
					resultset = append(resultset, c.Get("ID"))
				case "Door":
					resultset = append(resultset, k)
				case "Door.Delay":
					resultset = append(resultset, c.Get(fmt.Sprintf("Door[%v].Delay", oid)))
				}
				break
			}
		}
	}

	return resultset
}
