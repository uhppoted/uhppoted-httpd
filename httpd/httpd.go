package httpd

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

type HTTPD struct {
	Dir string
}

type dispatcher struct {
	root string
	fs   http.Handler
}

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	d := dispatcher{
		root: h.Dir,
		fs:   http.FileServer(fs),
	}

	srv := http.Server{
		Addr: ":8080",
	}

	shutdown := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("WARN  HTTP server shutdown error: %v", err)
		}

		close(shutdown)
	}()

	http.Handle("/", &d)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ERROR: %v", err)
	}

	<-shutdown
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	debug(fmt.Sprintf("%v", r.URL))

	authorized := false
	if cookie, err := r.Cookie("uhppoted-httpd-auth"); err == nil {
		if err := auth.Verify(cookie.Value); err != nil {
			info(err.Error())
		} else {
			authorized = true
		}
	}

	switch strings.ToUpper(r.Method) {
	case http.MethodGet:
		d.get(w, r, authorized)
	case http.MethodPost:
		d.post(w, r, authorized)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}

}

func (d *dispatcher) get(w http.ResponseWriter, r *http.Request, authorized bool) {
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if strings.HasSuffix(path, ".html") {
		var file string

		if authorized || path == "/login.html" {
			file = filepath.Clean(filepath.Join(d.root, path[1:]))
			getPage(file, w)
			return
		}

		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	d.fs.ServeHTTP(w, r)
}

func (d *dispatcher) post(w http.ResponseWriter, r *http.Request, authorized bool) {
	path := r.URL.Path

	if path == "/authorize" {
		r.ParseForm()

		uid := ""
		if v, ok := r.Form["uid"]; ok && len(v) > 0 {
			uid = v[0]
		}

		pwd := ""
		if v, ok := r.Form["pwd"]; ok && len(v) > 0 {
			pwd = v[0]
		}

		token, err := auth.Authorize(uid, pwd)
		if err != nil {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		cookie := http.Cookie{
			Name:     "uhppoted-httpd-auth",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(5 * time.Minute),
			MaxAge:   int((10 * time.Minute).Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			// Secure:   true,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	} else {
		http.Error(w, "NOT IMPLEMENTED", http.StatusNotImplemented)
	}
}

func getPage(file string, w http.ResponseWriter) {
	// TODO verify file is in a subdirectory
	// TODO igore . paths

	translate(file, w)
}

func authorize(header []string) error {
	if len(header) == 0 {
		return fmt.Errorf("Empty 'Authorization' header")
	}

	return nil
}

func translate(filename string, w http.ResponseWriter) {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	translation := filepath.Join("translations", "en", base+".json")

	bytes, err := ioutil.ReadFile(translation)
	if err != nil {
		warn(fmt.Sprintf("Error reading translation '%s'", translation), err)
		http.Error(w, "Gone Missing It Has", http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles(filename)
	if err != nil {
		warn(fmt.Sprintf("Error parsing template '%s'", filename), err)
		http.Error(w, "Sadly, All The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	page := map[string]interface{}{}

	err = json.Unmarshal(bytes, &page)
	if err != nil {
		warn(fmt.Sprintf("Error unmarshalling translation '%s')", translation), err)
		http.Error(w, "Sadly, Some Of The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	t.Execute(w, &page)
}

type httpdFileSystem struct {
	http.FileSystem
}

func (fs httpdFileSystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return nil, os.ErrPermission
		}
	}

	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	return httpdFile{file}, err
}

type httpdFile struct {
	http.File
}

func (f httpdFile) Readdir(n int) (fis []os.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}

	return
}

func debug(message string) {
	log.Printf("%-5s %s", "DEBUG", message)
}

func info(message string) {
	log.Printf("%-5s %s", "INFO", message)
}

func warn(message string, err error) {
	if err == nil {
		log.Printf("%-5s %s", "WARN", message)
	} else {
		log.Printf("%-5s %s", "WARN", message)
		log.Printf("%-5s %v", "", err)
	}
}
