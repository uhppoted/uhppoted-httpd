package httpd

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/cards"
	"github.com/uhppoted/uhppoted-httpd/httpd/controllers"
	"github.com/uhppoted/uhppoted-httpd/httpd/doors"
	"github.com/uhppoted/uhppoted-httpd/httpd/events"
	"github.com/uhppoted/uhppoted-httpd/httpd/groups"
	"github.com/uhppoted/uhppoted-httpd/httpd/interfaces"
	"github.com/uhppoted/uhppoted-httpd/httpd/logs"
	"github.com/uhppoted/uhppoted-httpd/httpd/users"
)

type handler struct {
	get  func(uid, role string, r *http.Request) interface{}
	post func(uid, role string, objects map[string]interface{}) (interface{}, error)
}

func (d *dispatcher) vtable(path string) *handler {
	switch path {
	case "/interfaces":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return interfaces.Get(uid, role) },
			post: interfaces.Post,
		}

	case "/controllers":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return controllers.Get(uid, role) },
			post: controllers.Post,
		}

	case "/doors":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return doors.Get(uid, role) },
			post: doors.Post,
		}

	case "/cards":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return cards.Get(uid, role) },
			post: cards.Post,
		}

	case "/groups":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return groups.Get(uid, role) },
			post: groups.Post,
		}

	case "/events":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return events.Get(uid, role, r) },
			post: nil,
		}

	case "/logs":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return logs.Get(uid, role, r) },
			post: nil,
		}

	case "/users":
		return &handler{
			get:  func(uid, role string, r *http.Request) interface{} { return users.Get(uid, role) },
			post: users.Post,
		}
	}

	return nil
}
