package controllers

import ()

type lanInfo struct {
	OID       string `json:"OID"`
	Interface string `json:"interface"`
	FieldName string `json:"field"`
	Current   string `json:"current"`
	Updated   string `json:"new"`
}

func (i lanInfo) Field() string {
	return i.FieldName
}

type controllerInfo struct {
	OID        string `json:"OID"`
	Controller string `json:"interface"`
	FieldName  string `json:"field"`
	Current    string `json:"current"`
	Updated    string `json:"new"`
}

func (i controllerInfo) Field() string {
	return i.FieldName
}
