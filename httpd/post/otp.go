package post

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
)

func VerifyOTP(w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	var uid string
	var pwd string

	contentType, acceptsGzip := parseHeader(r)

	if body, err := parseRequest(r, contentType); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else if uid, err = get(body, "uid"); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	} else if pwd, err = get(body, "pwd"); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	if err := auth.Verify(uid, pwd); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid user ID or password", http.StatusBadRequest)
		return
	}

	if acceptsGzip {

	}

	http.Error(w, "(work in progress)", http.StatusInternalServerError)
}
