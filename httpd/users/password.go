package users

import (
	"fmt"
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
)

func Password(uid, role string, w http.ResponseWriter, r *http.Request, auth auth.IAuth) (any, error) {
	if err := verifyAuthHeader(uid, r, auth); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return nil, err
	}

	if vars, err := getvars(r, "password"); err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return nil, err
	} else if pwd, ok := vars["password"]; !ok {
		http.Error(w, "Error in request", http.StatusBadRequest)
		return nil, fmt.Errorf("Missing password field")
	} else if err := system.SetPassword(uid, role, pwd); err != nil {
		return nil, err
	}

	return struct{}{}, nil
}
