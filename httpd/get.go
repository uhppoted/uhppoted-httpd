package httpd

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

const GZIP_MINIMUM = 16384

func (d *dispatcher) get(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != http.MethodGet {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	switch path {
	case "/interfaces",
		"/controllers",
		"/doors",
		"/cards",
		"/groups",
		"/events",
		"/logs",
		"/users":
		if handler := d.vtable(path); handler != nil && handler.get != nil {
			d.fetch(r, w, *handler)
		}

		return
	}

	d.getWithAuth(w, r)
}

func (d *dispatcher) getNoAuth(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != http.MethodGet {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	// ... parse headers, etc
	acceptsGzip := parseHeader(r)
	context := map[string]interface{}{
		"Theme": parseSettings(r),
		"Mode":  d.mode,
	}

	// ... normalise path
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	authorised := map[string]bool{}

	d.translate(path, context, authorised, w, acceptsGzip)
}

func (d *dispatcher) getWithAuth(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != http.MethodGet {
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
		return
	}

	// ... normalise path
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !strings.HasSuffix(path, ".html") {
		http.Error(w, fmt.Sprintf("No resource matching %v", r.URL.Path), http.StatusNotFound)
		return
	}

	// ... authenticated and authorized?
	uid, role, authenticated := d.authenticated(r, w)
	if !authenticated {
		d.unauthenticated(r, w)
		return
	}

	if ok := d.authorised(uid, role, path); !ok {
		d.unauthorised(r, w)
		return
	}

	// ... good to go
	acceptsGzip := parseHeader(r)
	context := map[string]interface{}{
		"Theme": parseSettings(r),
		"Mode":  d.mode,
		"User":  uid,
	}

	authorised := map[string]bool{
		"/sys/controllers.html": true,
		"/sys/doors.html":       true,
		"/sys/cards.html":       true,
		"/sys/groups.html":      true,
		"/sys/events.html":      true,
		"/sys/logs.html":        true,
		"/sys/users.html":       false,
	}

	for path, _ := range authorised {
		authorised[path] = d.authorised(uid, role, path)
	}

	d.translate(path, context, authorised, w, acceptsGzip)
}

func (d *dispatcher) translate(file string, context map[string]interface{}, authorised map[string]bool, w http.ResponseWriter, acceptsGzip bool) {
	type nav struct {
		Overview bool
		System   bool
		Doors    bool
		Cards    bool
		Groups   bool
		Events   bool
		Logs     bool
		Users    bool
	}

	page := map[string]interface{}{}

	page["context"] = context
	page["schema"] = schema.GetSchema()
	page["mode"] = ""
	page["readonly"] = false

	if d.mode == types.Monitor {
		page["readonly"] = true
	}

	// For a FS, use path.Join rather than filepath.Join (ref. https://pkg.go.dev/io/fs#ValidPath)
	translation := path.Join("translations", "en", strings.TrimSuffix(path.Base(file), path.Ext(file))+".json")

	if info, err := fs.Stat(d.fs, translation); err != nil {
		if !os.IsNotExist(err) {
			warn("", fmt.Errorf("Error locating translation '%s' (%w)", translation, err))
			http.Error(w, "Sadly, Most Of The Wheels All Came Off", http.StatusInternalServerError)
			return
		}
	} else if !info.IsDir() {
		if replacements, err := fs.ReadFile(d.fs, translation); err != nil {
			warn("", fmt.Errorf("Error reading translation '%s' (%w)", translation, err))
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		} else if err := json.Unmarshal(replacements, &page); err != nil {
			warn("", fmt.Errorf("Error unmarshalling translation '%s' (%w)", translation, err))
			http.Error(w, "Sadly, Some Of The Wheels All Came Off", http.StatusInternalServerError)
			return
		}
	}

	functions := template.FuncMap{
		"suffix": func(v string) string {
			tokens := strings.Split(v, ".")
			if len(tokens) > 0 {
				return tokens[len(tokens)-1]
			}

			return v
		},

		"nav": func(page string) interface{} {
			return struct {
				Page       string
				Authorised nav
			}{
				Page: page,
				Authorised: nav{
					Overview: authorised["/sys/overview.html"],
					System:   authorised["/sys/controllers.html"],
					Doors:    authorised["/sys/doors.html"],
					Cards:    authorised["/sys/cards.html"],
					Groups:   authorised["/sys/groups.html"],
					Events:   authorised["/sys/events.html"],
					Logs:     authorised["/sys/logs.html"],
					Users:    authorised["/sys/users.html"],
				},
			}
		},
	}

	// Ref. https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
	var name = path.Base(file)
	var filename = file

	if strings.HasPrefix(filename, "/") {
		filename = filename[1:]
	}

	t, err := template.New(name).Funcs(functions).ParseFS(d.fs, "templates/snippets.html", filename)
	if err != nil {
		warn("", fmt.Errorf("Error parsing template '%s' (%w)", file, err))
		http.Error(w, "Sadly, All The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	var b bytes.Buffer
	if err := t.Execute(&b, page); err != nil {
		warn("", fmt.Errorf("Error formatting page '%s' (%w)", file, err))
		http.Error(w, "Error formatting page", http.StatusInternalServerError)
		return
	}

	if acceptsGzip && b.Len() > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gz.Write(b.Bytes())
		gz.Close()
	} else {
		w.Write(b.Bytes())
	}
}

func (d *dispatcher) fetch(r *http.Request, w http.ResponseWriter, h handler) {
	path, err := resolve(r.URL)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	acceptsGzip := parseHeader(r)

	// ... authenticated and authorised?
	uid, role, ok := d.authenticated(r, w)
	if !ok {
		d.unauthenticated(r, w)
		return
	}

	// Returns empty JSON object if not authorised for the resource because this request may be
	// a legitimate part of the user interface.
	if ok := d.authorised(uid, role, path); !ok {
		if b, err := json.Marshal(struct{}{}); err != nil {
			http.Error(w, "Error generating response", http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		}

		return
	}

	// ... ok
	ctx, cancel := context.WithTimeout(d.context, d.timeout)

	defer cancel()

	var response interface{}

	go func() {
		response = h.get(uid, role, r)
		cancel()
	}()

	<-ctx.Done()

	if err := ctx.Err(); err != context.Canceled {
		warn("", err)
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error generating response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if acceptsGzip && len(b) > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gz.Write(b)
		gz.Close()
	} else {
		w.Write(b)
	}
}

func parseHeader(r *http.Request) bool {
	acceptsGzip := false

	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "accept-encoding" {
			for _, v := range h {
				if strings.Contains(strings.TrimSpace(strings.ToLower(v)), "gzip") {
					acceptsGzip = true
				}
			}
		}
	}

	return acceptsGzip
}

func parseSettings(r *http.Request) string {
	theme := "default"

	if cookie, err := r.Cookie(SettingsCookie); err == nil {
		re := regexp.MustCompile("(.*?):(.+)")
		tokens := strings.Split(cookie.Value, ",")
		for _, token := range tokens {
			match := re.FindStringSubmatch(token)
			if len(match) > 2 {
				if strings.TrimSpace(match[1]) == "theme" {
					theme = strings.TrimSpace(match[2])
				}
			}
		}
	}

	return theme
}
