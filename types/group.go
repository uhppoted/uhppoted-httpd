package types

type Group struct {
	ID    string
	Name  string
	Doors []string
}

type Groups []*Group

func (g *Group) Copy() *Group {
	replicant := Group{
		ID:    g.ID,
		Name:  g.Name,
		Doors: make([]string, len(g.Doors)),
	}

	copy(replicant.Doors, g.Doors)

	return &replicant
}

func (g *Groups) Copy() Groups {
	if g != nil {
		groups := make([]*Group, len(*g))

		for i, v := range *g {
			groups[i] = v.Copy()
		}

		return groups
	}

	return Groups{}
}
