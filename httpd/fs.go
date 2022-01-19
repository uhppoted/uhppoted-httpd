package httpd

import (
	"net/http"
	"os"
	"strings"
)

type httpdFileSystem struct {
	http.FileSystem
}

type httpdFile struct {
	http.File
}

func (f httpdFile) Readdir(n int) (fis []os.FileInfo, err error) {
	return nil, os.ErrPermission
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
