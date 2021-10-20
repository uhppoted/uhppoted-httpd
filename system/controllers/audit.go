package controllers

import ()

type lanInfo struct {
	Interface     string `json:"interface"`
	InterfaceName string `json:"name"`
	FieldName     string `json:"field"`
	Description   string `json:"description"`
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
	return i.Description
}

type controllerInfo struct {
	DeviceID    string `json:"device-id"`
	DeviceName  string `json:"device-name"`
	FieldName   string `json:"field"`
	Description string `json:"description"`
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
	return i.Description
}
