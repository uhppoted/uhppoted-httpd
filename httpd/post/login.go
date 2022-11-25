package post

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
)

func Login(w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	var uid string
	var pwd string

	if vars, err := get(r, "uid", "pwd"); err != nil {
		warnf("LOGIN", "%v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else {
		uid = vars["uid"]
		pwd = vars["pwd"]
	}

	loginCookie, err := r.Cookie(cookies.LoginCookie)
	if err != nil {
		warnf("LOGIN", "%v", err)
		cookies.Clear(w, cookies.SessionCookie, cookies.OTPCookie)
		http.Redirect(w, r, "/sys/login.html", http.StatusFound)
		return
	}

	if loginCookie == nil {
		warnf("LOGIN", "Missing login cookie")
		http.Error(w, "Missing login cookie", http.StatusBadRequest)
		return
	}

	sessionCookie, err := auth.Authenticate(uid, pwd, loginCookie)
	if err != nil {
		warnf("LOGIN", "%v", err)
		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	if sessionCookie != nil {
		http.SetCookie(w, sessionCookie)
	}

	cookies.Clear(w, cookies.LoginCookie)
}
