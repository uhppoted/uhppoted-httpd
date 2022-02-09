package httpd

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/auth"
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
	get  func(r *http.Request, a auth.OpAuth) interface{}
	post func(map[string]interface{}, auth.OpAuth) (interface{}, error)
}

func (d *dispatcher) vtable(path string) *handler {
	switch path {
	case "/interfaces":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return interfaces.Get(auth) },
			post: interfaces.Post,
		}

	case "/controllers":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return controllers.Get(auth) },
			post: controllers.Post,
		}

	case "/doors":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return doors.Get(auth) },
			post: doors.Post,
		}

	case "/cards":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return cards.Get(auth) },
			post: cards.Post,
		}

	case "/groups":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return groups.Get(auth) },
			post: groups.Post,
		}

	case "/events":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return events.Get(r, auth) },
			post: nil,
		}

	case "/logs":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return logs.Get(r, auth) },
			post: nil,
		}

	case "/users":
		return &handler{
			get:  func(r *http.Request, auth auth.OpAuth) interface{} { return users.Get(auth) },
			post: nil,
		}
	}

	return nil
}
