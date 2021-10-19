package controllers

import (
	"fmt"
)

type lanInfo struct {
	OID       string `json:"OID"`
	Interface string `json:"interface"`
	FieldName string `json:"field"`
	Current   string `json:"current"`
	Updated   string `json:"new"`
}

func (i lanInfo) ID() string {
	return ""
}

func (i lanInfo) Name() string {
	return i.Interface
}

func (i lanInfo) Field() string {
	return i.FieldName
}

func (i lanInfo) Details() string {
	return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
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
