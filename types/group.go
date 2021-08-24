package types

type Group struct {
	OID  string
	Name string
}

type Groups map[string]Group

func (g *Group) Clone() Group {
	return Group{
		OID:  g.OID,
		Name: g.Name,
	}
}

func (g *Groups) Clone() Groups {
	if g != nil {
		groups := map[string]Group{}

		for gid, v := range *g {
			groups[gid] = v.Clone()
		}

		return groups
	}

	return Groups{}
}
