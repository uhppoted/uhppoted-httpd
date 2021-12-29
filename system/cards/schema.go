package cards

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const CardStatus = catalog.Status
const CardCreated = catalog.Created
const CardDeleted = catalog.Deleted
const CardModified = catalog.Modified
const CardName = catalog.CardName
const CardNumber = catalog.CardNumber
const CardFrom = catalog.CardFrom
const CardTo = catalog.CardTo
const CardGroups = catalog.CardGroups
const GroupName = catalog.GroupName

var lookup = map[catalog.Suffix]string{
	CardStatus:   "card.status",
	CardCreated:  "card.created",
	CardDeleted:  "card.deleted",
	CardModified: "card.modified",
	CardName:     "card.name",
	CardNumber:   "card.number",
	CardFrom:     "card.from",
	CardTo:       "card.to",
	CardGroups:   "card.groups",
}
