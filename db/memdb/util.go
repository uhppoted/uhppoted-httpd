package memdb

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type object map[string]interface{}

func (o object) name() (string, error) {
	if s, ok := o["name"].(string); ok {
		return strings.TrimSpace(s), nil
	}

	return "", &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid card holder name"),
		Detail: fmt.Errorf("Invalid card holder name: %v", o["name"]),
	}
}

func (o object) card() (string, error) {
	if s, ok := o["card"].(string); ok {
		return strings.TrimSpace(s), nil
	}

	if u, ok := o["card"].(uint32); ok {
		return strconv.FormatUint(uint64(u), 10), nil
	}

	return "", &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid card number"),
		Detail: fmt.Errorf("Invalid card number: %v", o["card"]),
	}
}

func (o object) from() (string, error) {
	if s, ok := o["from"].(string); ok {
		if s == "" {
			return s, nil
		}

		if v, err := time.Parse("2006-01-02", s); err == nil {
			return v.Format("2006-01-02"), nil
		}
	}

	return "", &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid 'from' date"),
		Detail: fmt.Errorf("Invalid 'from' date: %v", o["from"]),
	}
}

func (o object) to() (string, error) {
	if s, ok := o["to"].(string); ok {
		if s == "" {
			return s, nil
		}

		if v, err := time.Parse("2006-01-02", s); err == nil {
			return v.Format("2006-01-02"), nil
		}
	}

	return "", &types.HttpdError{
		Status: http.StatusBadRequest,
		Err:    fmt.Errorf("Invalid 'to' date"),
		Detail: fmt.Errorf("Invalid 'to' date: %v", o["to"]),
	}
}
