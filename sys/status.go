package system

type status int

const (
	StatusUnknown status = iota
	StatusOk
	StatusUncertain
)

func (s status) String() string {
	return [...]string{"unknown", "ok", "uncertain"}[s]
}
