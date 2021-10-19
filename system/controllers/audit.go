package controllers

import (
	"fmt"
	"strings"
)

type lanInfo struct {
	Interface     string `json:"interface"`
	InterfaceName string `json:"name"`
	FieldName     string `json:"field"`
	Current       string `json:"current"`
	Updated       string `json:"new"`
}

func (i lanInfo) ID() string {
	return i.Interface
}

func (i lanInfo) Name() string {
	return i.InterfaceName
}

func (i lanInfo) Field() string {
	return i.FieldName
}

func (i lanInfo) Details() string {
	switch strings.ToLower(i.FieldName) {
	case "name":
		return fmt.Sprintf("Updated name from %v to %v", i.Current, i.Updated)

	case "bind":
		return fmt.Sprintf("Updated bind address from %v to %v", i.Current, i.Updated)

	case "broadcast":
		return fmt.Sprintf("Updated broadcast address from %v to %v", i.Current, i.Updated)

	case "listen":
		return fmt.Sprintf("Updated listen address from %v to %v", i.Current, i.Updated)

	default:
		return fmt.Sprintf("Updated %v from %v to %v", i.FieldName, i.Current, i.Updated)
	}
}

type controllerInfo struct {
	OID        string `json:"OID"`
	DeviceID   string `json:"device-id"`
	DeviceName string `json:"device-name"`
	FieldName  string `json:"field"`
	Current    string `json:"current"`
	Updated    string `json:"new"`
}

func (i controllerInfo) ID() string {
	return i.DeviceID
}

func (i controllerInfo) Name() string {
	return i.DeviceName
}

func (i controllerInfo) Field() string {
	return i.FieldName
}

func (i controllerInfo) Details() string {
	return fmt.Sprintf("Updated %v from %v to %v", i.FieldName, i.Current, i.Updated)
}
