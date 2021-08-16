package system

import (
	"regexp"
	"strings"
)

type Resolver struct {
}

func (r Resolver) Get(query string) []interface{} {
	q := strings.ToLower(query)

	switch {
	case strings.HasPrefix(q, "controller.oid for door.oid"):
		return r.lookupControllerForDoor(q)

	case strings.HasPrefix(q, "controller.created for door.oid"):
		return r.lookupControllerForDoor(q)

	case strings.HasPrefix(q, "controller.name for door.oid"):
		return r.lookupControllerForDoor(q)

	case strings.HasPrefix(q, "controller.id for door.oid"):
		return r.lookupControllerForDoor(q)

	case strings.HasPrefix(q, "controller.door for door.oid"):
		return r.lookupControllerForDoor(q)

	case strings.HasPrefix(q, "door.delay.configured for door.oid"):
		return r.lookupDoor(q)

	case strings.HasPrefix(q, "door.mode.configured for door.oid"):
		return r.lookupDoor(q)
	}

	return nil
}

func (r Resolver) lookupControllerForDoor(query string) []interface{} {
	re := regexp.MustCompile(`controller\.(oid|created|name|id|door|door\.mode|door\.delay|door\.delay.dirty|door\.control\.dirty) for door\.oid\[(.*?)\]`)

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
				case "oid":
					resultset = append(resultset, c.Get("oid"))
				case "created":
					resultset = append(resultset, c.Get("created"))
				case "name":
					resultset = append(resultset, c.Get("name"))
				case "id":
					resultset = append(resultset, c.Get("ID"))
				case "door":
					resultset = append(resultset, k)
				}
				break
			}
		}
	}

	return resultset
}

func (r Resolver) lookupDoor(query string) []interface{} {
	re := regexp.MustCompile(`door\.(delay|delay\.configured|mode|mode\.configured) for door\.oid\[(.*?)\]`)
	match := re.FindStringSubmatch(query)

	if match == nil || len(match) < 3 {
		return nil
	}

	field := match[1]
	oid := match[2]

	if door, ok := sys.doors.Doors[oid]; ok {
		switch field {
		case "delay.configured":
			return []interface{}{door.Get("delay.configured")}

		case "mode.configured":
			return []interface{}{door.Get("Mode.configured")}
		}
	}

	return nil
}

// func (r Resolver) get(query string) []interface{} {
// 	re := regexp.MustCompile(`controller\.(oid|created|name|id|door) for door\.oid\[(.*?)\]`)
//
// 	match := re.FindStringSubmatch(query)
// 	if match == nil || len(match) < 3 {
// 		return nil
// 	}
//
// 	field := match[1]
// 	oid := match[2]
// 	resultset := []interface{}{}
//
// 	for _, c := range sys.controllers.Controllers {
// 		for _, d := range c.Doors {
// 			if d == oid {
// 				switch field {
// 				//				case "oid":
// 				//					resultset = append(resultset, c.Get("oid"))
// 				//				case "created":
// 				//					resultset = append(resultset, c.Get("created"))
// 				//				case "name":
// 				//					resultset = append(resultset, c.Get("name"))
// 				//				case "id":
// 				//					resultset = append(resultset, c.Get("ID"))
// 				//				case "door":
// 				//					resultset = append(resultset, k)
// 				}
// 				break
// 			}
// 		}
// 	}
//
// 	return resultset
// }
