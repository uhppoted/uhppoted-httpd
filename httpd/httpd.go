package httpd

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type HTTPD struct {
	Dir string
}

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	http.Handle("/", http.FileServer(fs))

	log.Fatal(http.ListenAndServe(":8080", nil))
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
