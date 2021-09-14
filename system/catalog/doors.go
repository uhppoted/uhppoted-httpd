package catalog

func Doors() []OID {
	list := []OID{}

	for d, _ := range catalog.doors {
		list = append(list, d)
	}

	return list
}
