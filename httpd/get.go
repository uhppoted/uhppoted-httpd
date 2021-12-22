package httpd

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/cards"
	"github.com/uhppoted/uhppoted-httpd/httpd/controllers"
	"github.com/uhppoted/uhppoted-httpd/httpd/doors"
	"github.com/uhppoted/uhppoted-httpd/httpd/events"
	"github.com/uhppoted/uhppoted-httpd/httpd/groups"
	"github.com/uhppoted/uhppoted-httpd/httpd/interfaces"
	"github.com/uhppoted/uhppoted-httpd/httpd/logs"
	"github.com/uhppoted/uhppoted-httpd/system"
)

const GZIP_MINIMUM = 16384

func (d *dispatcher) get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Authorise unless images,CSS,etc
	prefixes := []string{"/images/", "/css/", "/javascript/"}
	files := []string{"/manifest.json"}
	authorised := false

	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			authorised = true
		}
	}

	for _, f := range files {
		if path == f {
			authorised = true
		}
	}

	if !authorised {
		if _, _, ok := d.authorized(w, r, path); !ok {
			return
		}
	}

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

	switch path {
	case "/interfaces":
		interfaces.Fetch(w, r, d.timeout)
		return

	case "/controllers":
		controllers.Fetch(w, r, d.timeout)
		return

	case "/doors":
		doors.Fetch(w, r, d.timeout)
		return

	case "/cards":
		cards.Fetch(w, r, d.timeout)
		return

	case "/groups":
		groups.Fetch(w, r, d.timeout)
		return

	case "/events":
		events.Fetch(w, r, d.timeout)
		return

	case "/logs":
		logs.Fetch(w, r, d.timeout)
		return
	}

	if strings.HasSuffix(path, ".html") {
		file := filepath.Clean(filepath.Join(d.root, path[1:]))
		context := map[string]interface{}{
			"User":  d.user(r),
			"Theme": theme,
		}

		d.translate(file, context, w, acceptsGzip)
		return
	}

	d.fs.ServeHTTP(w, r)
}

func (d *dispatcher) translate(filename string, context map[string]interface{}, w http.ResponseWriter, acceptsGzip bool) {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	translation := filepath.Join("translations", "en", base+".json")
	page := map[string]interface{}{}

	page["context"] = context
	page["system"] = system.System()
	page["schema"] = system.Schema()

	info, err := os.Stat(translation)
	if err != nil && !os.IsNotExist(err) {
		warn(fmt.Errorf("Error locating translation '%s' (%w)", translation, err))
		http.Error(w, "Sadly, Most Of The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	if err == nil && !info.IsDir() {
		replacements, err := ioutil.ReadFile(translation)
		if err != nil {
			warn(fmt.Errorf("Error reading translation '%s' (%w)", translation, err))
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		err = json.Unmarshal(replacements, &page)
		if err != nil {
			warn(fmt.Errorf("Error unmarshalling translation '%s' (%w)", translation, err))
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
	}

	// Ref. https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
	snippets := "html/templates/snippets.html"
	name := filepath.Base(filename)
	t, err := template.New(name).Funcs(functions).ParseFiles(snippets, filename)
	if err != nil {
		warn(fmt.Errorf("Error parsing template '%s' (%w)", filename, err))
		http.Error(w, "Sadly, All The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	var b bytes.Buffer
	if err := t.Execute(&b, &page); err != nil {
		warn(fmt.Errorf("Error formatting page '%s' (%w)", filename, err))
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
