package catalog

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

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

type CatalogInterface struct {
	OID schema.OID
}

func (t CatalogInterface) TypeOf() Type {
	return TInterface
}

func (t CatalogInterface) oid() schema.OID {
	return t.OID
}

type CatalogController struct {
	OID      schema.OID
	DeviceID uint32
}

func (t CatalogController) TypeOf() Type {
	return TController
}

func (t CatalogController) oid() schema.OID {
	return t.OID
}

type CatalogDoor struct {
	OID schema.OID
}

func (t CatalogDoor) TypeOf() Type {
	return TDoor
}

func (t CatalogDoor) oid() schema.OID {
	return t.OID
}

type CatalogCard struct {
	OID schema.OID
}

func (t CatalogCard) TypeOf() Type {
	return TCard
}

func (t CatalogCard) oid() schema.OID {
	return t.OID
}

type CatalogGroup struct {
	OID schema.OID
}

func (t CatalogGroup) TypeOf() Type {
	return TGroup
}

func (t CatalogGroup) oid() schema.OID {
	return t.OID
}

type CatalogEvent struct {
	OID      schema.OID
	DeviceID uint32
	Index    uint32
}

func (t CatalogEvent) TypeOf() Type {
	return TEvent
}

func (t CatalogEvent) oid() schema.OID {
	return t.OID
}

type CatalogLogEntry struct {
	OID schema.OID
}

func (t CatalogLogEntry) TypeOf() Type {
	return TLogEntry
}

func (t CatalogLogEntry) oid() schema.OID {
	return t.OID
}

type CatalogUser struct {
	OID schema.OID
}

func (t CatalogUser) TypeOf() Type {
	return TUser
}

func (t CatalogUser) oid() schema.OID {
	return t.OID
}
