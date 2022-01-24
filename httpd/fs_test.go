package httpd

import (
	"net/http"
	"os"
	"testing"
)

func TestFSDotFileHidingWithDir(t *testing.T) {
	fs := filesystem{
		http.Dir("."),
	}

	f, err := fs.Open(".fs_test.html")
	if err == nil {
		t.Errorf("Expected error 'permission denied', got:%v", err)
	}

	if err != os.ErrPermission {
		t.Errorf("Expected error 'permission denied', got:%v", err)
	}

	if f != nil {
		t.Errorf("Expected 'nil' file, got:%v", f)
	}
}

func TestFSDotFileHidingWithDirFS(t *testing.T) {
	fs := filesystem{
		http.FS(os.DirFS(".")),
	}

	f, err := fs.Open(".fs_test.html")
	if err == nil {
		t.Errorf("Expected error 'permission denied', got:%v", err)
	}

	if err != os.ErrPermission {
		t.Errorf("Expected error 'permission denied', got:%v", err)
	}

	if f != nil {
		t.Errorf("Expected 'nil' file, got:%v", f)
	}
}

func TestFSReadDirWithDir(t *testing.T) {
	fs := filesystem{
		http.Dir("."),
	}

	f, err := fs.Open("/")
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if f == nil {
		t.Errorf("Expected valid file handle, got:%v", f)
	}

	if info, err := f.Readdir(1); err == nil || info != nil {
		if err == nil {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if err != os.ErrPermission {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if info != nil {
			t.Errorf("Expected 'nil' file info, got:%v", info)
		}
	}

	if info, err := f.Readdir(-1); err == nil || info != nil {
		if err == nil {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if err != os.ErrPermission {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if info != nil {
			t.Errorf("Expected 'nil' file info, got:%v", info)
		}
	}
}

func TestFSReadDirWithDirFS(t *testing.T) {
	fs := filesystem{
		http.FS(os.DirFS(".")),
	}

	f, err := fs.Open("/")
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if f == nil {
		t.Errorf("Expected valid file handle, got:%v", f)
	}

	if info, err := f.Readdir(1); err == nil || info != nil {
		if err == nil {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if err != os.ErrPermission {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if info != nil {
			t.Errorf("Expected 'nil' file info, got:%v", info)
		}
	}

	if info, err := f.Readdir(-1); err == nil || info != nil {
		if err == nil {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if err != os.ErrPermission {
			t.Errorf("Expected error 'permission denied', got:%v", err)
		}

		if info != nil {
			t.Errorf("Expected 'nil' file info, got:%v", info)
		}
	}
}
