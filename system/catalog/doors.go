package catalog

func GetDoors() []OID {
	list := []OID{}

	for d, _ := range catalog.doors {
		list = append(list, d)
	}

	return list
}
