package httpd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

func (d *dispatcher) get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !d.authorized(w, r, path) {
		return
	}

	if strings.HasSuffix(path, ".html") {
		context := map[string]interface{}{}
		context["User"] = d.user(r)
		file := filepath.Clean(filepath.Join(d.root, path[1:]))

		translate(file, context, w)
		return
	}

	d.fs.ServeHTTP(w, r)
}

func translate(filename string, context map[string]interface{}, w http.ResponseWriter) {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	translation := filepath.Join("translations", "en", base+".json")

	html, err := ioutil.ReadFile(translation)
	if err != nil {
		warn(fmt.Errorf("Error reading translation '%s' (%w)", translation, err))
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles(filename)
	if err != nil {
		warn(fmt.Errorf("Error parsing template '%s' (%w)", filename, err))
		http.Error(w, "Sadly, All The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	page := map[string]interface{}{}

	err = json.Unmarshal(html, &page)
	if err != nil {
		warn(fmt.Errorf("Error unmarshalling translation '%s' (%w)", translation, err))
		http.Error(w, "Sadly, Some Of The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	page["context"] = context
	page["db"] = NewDB()

	var s strings.Builder

	if err := t.Execute(&s, &page); err != nil {
		warn(fmt.Errorf("Error formatting page '%s' (%w)", filename, err))
		http.Error(w, "Error formatting page", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", s.String())
}
