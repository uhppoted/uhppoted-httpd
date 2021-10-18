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

func (i lanInfo) Field() string {
	return i.FieldName
}

func (i lanInfo) Details() string {
	return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
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

func (i controllerInfo) Details() string {
	return fmt.Sprintf("updated %v from %v to %v", i.FieldName, i.Current, i.Updated)
}
