package catalog

func Groups() []OID {
	list := []OID{}

	for g, _ := range catalog.groups {
		list = append(list, g)
	}

	return list
}
