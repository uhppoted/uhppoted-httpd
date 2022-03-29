package catalog

type Type int

const (
	TInterface Type = iota
	TController
	TDoor
	TCard
	TGroup
	TEvent
	TLogEntry
	TUser
)

func (t Type) String() string {
	return []string{
		"interface",
		"controller",
		"door",
		"card",
		"group",
		"event",
		"log entry",
		"user",
	}[t]
}

// NTS: workaround for Go generics
//
// Ref. https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#methods-may-not-take-additional-type-arguments(

func (t CatalogInterface) TypeOf() Type {
	return TInterface
}

func (t CatalogController) TypeOf() Type {
	return TController
}

func (t CatalogDoor) TypeOf() Type {
	return TDoor
}

func (t CatalogCard) TypeOf() Type {
	return TCard
}

func (t CatalogGroup) TypeOf() Type {
	return TGroup
}

func (t CatalogEvent) TypeOf() Type {
	return TEvent
}

func (t CatalogLogEntry) TypeOf() Type {
	return TLogEntry
}

func (t CatalogUser) TypeOf() Type {
	return TUser
}
