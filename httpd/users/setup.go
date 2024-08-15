package users

import (
	"net/http"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Setup(w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	if vars, err := getvars(r, "name", "uid", "pwd"); err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
	} else if uid, ok := vars["uid"]; !ok || strings.TrimSpace(uid) == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
	} else if pwd, ok := vars["pwd"]; !ok {
		http.Error(w, "Missing password", http.StatusBadRequest)
	} else {
		name := ""
		if v, ok := vars["name"]; ok {
			name = strings.TrimSpace(v)
		}

		role := auth.AdminRole()

		if err := system.MakeAdminUser(name, uid, pwd, role); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Ok", http.StatusOK)
		}
	}
}
