package catalog

func GetGroups() []OID {
	list := []OID{}

	for g, _ := range catalog.groups {
		list = append(list, g)
	}

	return list
}
