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
)

type handler struct {
	tag   string
	rules string
	get   func(*http.Request, auth.OpAuth) interface{}
	post  func(map[string]interface{}, auth.OpAuth) (interface{}, error)
}

func (d *dispatcher) vtable(path string) *handler {
	switch path {
	case "/interfaces":
		return &handler{
			tag:   "system",
			rules: d.grule.system,
			get:   func(r *http.Request, a auth.OpAuth) interface{} { return interfaces.Get() },
			post:  interfaces.Post,
		}

	case "/controllers":
		return &handler{
			tag:   "system",
			rules: d.grule.system,
			get:   func(r *http.Request, a auth.OpAuth) interface{} { return controllers.Get() },
			post:  controllers.Post,
		}

	case "/doors":
		return &handler{
			tag:   "doors",
			rules: d.grule.doors,
			get:   func(r *http.Request, auth auth.OpAuth) interface{} { return doors.Get(auth) },
			post:  doors.Post,
		}

	case "/cards":
		return &handler{
			tag:   "cards",
			rules: d.grule.cards,
			get:   func(r *http.Request, auth auth.OpAuth) interface{} { return cards.Get(auth) },
			post:  cards.Post,
		}

	case "/groups":
		return &handler{
			tag:   "groups",
			rules: d.grule.groups,
			get:   func(r *http.Request, auth auth.OpAuth) interface{} { return groups.Get(auth) },
			post:  groups.Post,
		}

	case "/events":
		return &handler{
			tag:   "events",
			rules: "",
			get:   func(r *http.Request, a auth.OpAuth) interface{} { return events.Get(r) },
			post:  nil,
		}

	case "/logs":
		return &handler{
			tag:   "logs",
			rules: "",
			get:   func(r *http.Request, a auth.OpAuth) interface{} { return logs.Get(r) },
			post:  nil,
		}
	}

	return nil
}
