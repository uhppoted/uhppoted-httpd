package types

type Group struct {
	ID    string
	Name  string
	Doors []string
}

type Groups map[string]Group

func (g *Group) Clone() Group {
	replicant := Group{
		ID:    g.ID,
		Name:  g.Name,
		Doors: make([]string, len(g.Doors)),
	}

	copy(replicant.Doors, g.Doors)

	return replicant
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
