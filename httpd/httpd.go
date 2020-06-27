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
)

type HTTPD struct {
	Dir string
}

type dispatcher struct {
	fs http.Handler
}

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	d := dispatcher{
		fs: http.FileServer(fs),
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
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if strings.HasSuffix(path, ".html") {

		handlePage(path, w, r)
		return
	}

	d.fs.ServeHTTP(w, r)
}

func handlePage(path string, w http.ResponseWriter, r *http.Request) {
	dir := "html"
	filename := filepath.Join(dir, path[1:])
	translation := filepath.Join("translations", "en", "index.json")

	// TODO verify file is in a subdirectory
	// TODO igore . paths

	bytes, err := ioutil.ReadFile(translation)
	if err != nil {
		http.Error(w, "Gone Missing It Has", http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles(filename)
	if err != nil {
		warn(fmt.Sprintf("Error parsing template '%s' (%w)", filename, err))
		http.Error(w, "Sadly, All The Wheels All Came Off", http.StatusInternalServerError)
		return
	}

	page := map[string]interface{}{}

	err = json.Unmarshal(bytes, &page)
	if err != nil {
		warn(fmt.Sprintf("Error unmarshalling translation '%s' (%w)", translation, err))
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

func warn(message string) {
	log.Printf("%-5s %s", "WARN", message)
}
