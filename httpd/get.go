package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/cardholders"
)

func (d *dispatcher) get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if strings.HasSuffix(path, ".html") {
		_, ok := d.authorized(w, r, path)
		if !ok {
			return
		}
	}

	switch path {
	case "/cardholders":
		cardholders.Fetch(d.db, w, r, d.timeout)
		return
	}

	if strings.HasSuffix(path, ".html") {
		context := map[string]interface{}{}
		context["User"] = d.user(r)
		file := filepath.Clean(filepath.Join(d.root, path[1:]))

		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
		d.translate(file, context, w)
		return
	}

	d.fs.ServeHTTP(w, r)
}

func (d *dispatcher) translate(filename string, context map[string]interface{}, w http.ResponseWriter) {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	translation := filepath.Join("translations", "en", base+".json")
	page := map[string]interface{}{}

	page["context"] = context
	page["db"] = d.db

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

	t, err := template.ParseFiles(filename)
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

	w.Write(b.Bytes())
}
