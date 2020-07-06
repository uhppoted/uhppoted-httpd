module github.com/uhppoted/uhppoted-httpd

go 1.14

require (
	github.com/cristalhq/jwt/v3 v3.0.0
	github.com/google/uuid v1.1.1
	github.com/uhppoted/uhppoted-api v0.6.3
)

replace github.com/uhppoted/uhppoted-api => ../uhppoted-api
