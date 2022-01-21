package httpd

import (
	"net/http"
	"os"
	"strings"
)

type filesystem struct {
	http.FileSystem
}

func (fs filesystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return nil, os.ErrPermission
		}
	}

	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	return file{f}, err
}

type file struct {
	http.File
}

func (f file) Readdir(N int) (fis []os.FileInfo, err error) {
	return nil, os.ErrPermission
}
