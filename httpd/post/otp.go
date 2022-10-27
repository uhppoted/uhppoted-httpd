package post

import (
	"fmt"
	"net/http"
)

func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var uid string
	var pwd string

	contentType, acceptsGzip := parseHeader(r)

	if body, err := parseRequest(r, contentType); err != nil {
		warnf("OTP", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else if uid, err = get(body, "uid"); err != nil {
		warnf("OTP", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	} else if pwd, err = get(body, "pwd"); err != nil {
		warnf("OTP", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	if ok, err := validatePassword(uid, pwd); err != nil {
		warnf("OTP", err)
		http.Error(w, "Error validating password", http.StatusBadRequest)
		return
	} else if !ok {
		warnf("OTP", fmt.Errorf("invalid password"))
		http.Error(w, "Error validating password", http.StatusBadRequest)
		return
	}

	if acceptsGzip {

	}

	http.Error(w, "(work in progress)", http.StatusInternalServerError)
}

func validatePassword(uid, pwd string) (bool, error) {
	// if err := auth.Verify(uid, old); err != nil {
	// 	return nil, fmt.Errorf("Invalid user ID or password")
	// }

	// if err := system.SetPassword(uid, pwd); err != nil {
	// 	return nil, err
	// }

	return false, nil
}
