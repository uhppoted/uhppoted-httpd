package types

type RunMode int

const (
	Normal RunMode = iota
	Monitor
	Synchronize
)

func (r RunMode) String() string {
	return []string{"normal", "monitor", "synchronize"}[r]
}

func ParseRunMode(s string) RunMode {
	switch s {
	case "monitor":
		return Monitor

	case "synchronize":
		return Synchronize

	default:
		return Normal
	}
}
