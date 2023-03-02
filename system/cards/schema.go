package cards

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const CardStatus = schema.Status
const CardCreated = schema.Created
const CardDeleted = schema.Deleted
const CardModified = schema.Modified
const CardName = schema.CardName
const CardNumber = schema.CardNumber
const CardPIN = schema.CardPIN
const CardFrom = schema.CardFrom
const CardTo = schema.CardTo
const CardGroups = schema.CardGroups
const GroupName = schema.GroupName

var lookup = map[schema.Suffix]string{
	CardStatus:   "card.status",
	CardCreated:  "card.created",
	CardDeleted:  "card.deleted",
	CardModified: "card.modified",
	CardName:     "card.name",
	CardNumber:   "card.number",
	CardPIN:      "card.PIN",
	CardFrom:     "card.from",
	CardTo:       "card.to",
	CardGroups:   "card.groups",
}
