package httpd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type filesystem struct {
	http.FileSystem
}

func (fss filesystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return nil, os.ErrPermission
		}
	}

	f, err := fss.FileSystem.Open(name)
	fmt.Printf(">>> DEBUG/FS:Open/Y %v  %v\n", name, err)
	if err != nil {
		return nil, err
	}

	return file{f}, nil
}

type file struct {
	http.File
}

func (f file) Readdir(N int) (fis []os.FileInfo, err error) {
	return nil, os.ErrPermission
}
