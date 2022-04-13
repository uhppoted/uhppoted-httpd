package types

type Interval struct {
	From uint32
	To   uint32
}

func (i Interval) Contains(v uint32) bool {
	return i.From <= v && i.To >= v
}
