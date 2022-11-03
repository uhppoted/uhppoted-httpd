package cookies

import (
	"net/http"
)

const (
	SettingsCookie = "uhppoted-settings"
	LoginCookie    = "uhppoted-httpd-login"
	SessionCookie  = "uhppoted-httpd-session"
	OTPCookie      = "uhppoted-httpd-otp"
)

// cf. https://stackoverflow.com/questions/27671061/how-to-delete-cookie
func Clear(w http.ResponseWriter, cookies ...string) {
	for _, cookie := range cookies {
		http.SetCookie(w, &http.Cookie{
			Name:     cookie,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			//  Secure:   true,
		})
	}
}
