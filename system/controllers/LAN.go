package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
)

type LAN struct {
	interfaces interfaces.Interfaces
	lan        interfaces.LAN
}

func (l *LAN) compare(controllers []*Controller, permissions acl.ACL) error {
	devices := []types.IController{}
	for _, c := range controllers {
		if c.realized() {
			devices = append(devices, c.AsIController())
		}
	}

	return l.lan.CompareACL(devices, permissions)
}

func (l *LAN) update(controllers []*Controller, permissions acl.ACL) error {
	devices := []types.IController{}
	for _, c := range controllers {
		if c.realized() {
			devices = append(devices, c.AsIController())
		}
	}

	return l.lan.UpdateACL(devices, permissions)
}
