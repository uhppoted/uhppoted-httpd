package users

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/system"
)

func Setup(w http.ResponseWriter, r *http.Request) {
	if vars, err := getvars(r, "uid", "pwd"); err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
	} else if uid, ok := vars["uid"]; !ok {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
	} else if pwd, ok := vars["pwd"]; !ok {
		http.Error(w, "Missing password", http.StatusBadRequest)
	} else if err := system.MakeAdminUser(uid, pwd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, "Ok", http.StatusOK)
	}
}
