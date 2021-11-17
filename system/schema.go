package system

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type schema struct {
	Interfaces  interfaces  `json:"interfaces"`
	Controllers catalog.OID `json:"controllers"`
	Doors       catalog.OID `json:"doors"`
	Cards       catalog.OID `json:"cards"`
	Groups      catalog.OID `json:"groups"`
	Events      catalog.OID `json:"events"`
	Logs        catalog.OID `json:"logs"`
}

type interfaces struct {
	OID       catalog.OID    `json:"base"`
	Type      catalog.Suffix `json:"type"`
	Name      catalog.Suffix `json:"name"`
	Bind      catalog.Suffix `json:"bind"`
	Broadcast catalog.Suffix `json:"broadcast"`
	Listen    catalog.Suffix `json:"listen"`
}

func Schema() interface{} {
	return schema{
		Interfaces: interfaces{
			OID:       catalog.InterfacesOID,
			Type:      catalog.InterfaceType,
			Name:      catalog.InterfaceName,
			Bind:      catalog.LANBindAddress,
			Broadcast: catalog.LANBroadcastAddress,
			Listen:    catalog.LANListenAddress,
		},
		Controllers: catalog.ControllersOID,
		Doors:       catalog.DoorsOID,
		Cards:       catalog.CardsOID,
		Groups:      catalog.GroupsOID,
		Events:      catalog.EventsOID,
		Logs:        catalog.LogsOID,
	}
}
