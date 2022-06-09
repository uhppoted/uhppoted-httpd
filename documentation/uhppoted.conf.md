# `uhppoted.conf`

`uhppoted.conf` is the shared configuration file for all the `uhppoted` modules and is variously located in:

- /etc/uhppoted/uhppoted.conf (Linux)
- /usr/local/etc/com.github.uhppoted/uhppoted.conf (MacOS)
- \Program Data\uhppoted\uhppoted.conf (Windows)

The file contains optional configuration sections for all supported modules. For `uhppoted-httpd`, only 
the _HTTPD_ section is relevant.

## _HTTPD_

| *Attribute*                            | *Description*                                      | *Default*                          |
| -------------------------------------- | -------------------------------------------------- |----------------------------------- |
| httpd.html                             | Folder containing the HTML pages, images, etc      | Embedded HTML                      |
| httpd.http.enabled                     | Enables/disables the HTTP server i.e. without TLS  | `false`                            |
| httpd.http.port                        | HTTP server port                                   | 8080                               |
| httpd.https.enabled                    | Enables/disables the HTTPS server                  | `true`                             |
| httpd.https.port                       | HTTPS server port                                  | 8443                               |
| httpd.tls.ca                           | HTTPS server CA certificate PEM file               | _config_/httpd/ca.cert             |
| httpd.tls.certificate                  | HTTPS server TLS certificate PEM file              | _config_/httpd/uhppoted.cert       |
| httpd.tls.key                          | HTTPS server TLS key PEM file                      | _config_/httpd/uhppoted.key        |
| httpd.tls.client.certificates.required | Enforces client mutual TLS authentication          | `false`                            |
| httpd.security.auth                    | Authorization for HTTP requests (none/some)        | some                               |
| httpd.security.local.db                | auth.json file                                     | _config_/httpd/auth.json           |
| httpd.security.cookie.max-age          | Security cookie expiry (hours)                     | 24                                 |
| httpd.security.login.expiry            | Login cookie expiry e.g. 5m                        | 1m                                 |
| httpd.security.session.expiry          | Session cookie expiry e.g. 300s                    | 5m                                 |
| httpd.request.timeout                  | Time limit for fulfilling an HTTP request          | 15s                                |
| httpd.system.interfaces                | System file for data                               | _var_/system/interfaces.json       |
| httpd.system.controllers               | System file for data                               | _var_/system/controllers.json      |
| httpd.system.doors                     | System file for data                               | _var_/system/doors.json            |
| httpd.system.groups                    | System file for data                               | _var_/system/groups.json           |
| httpd.system.cards                     | System file for data                               | _var_/system/cards.json            |
| httpd.system.events                    | System file for data                               | _var_/system/events.json           |
| httpd.system.logs                      | System file for data                               | _var_/system/logs.json             |
| httpd.system.users                     | System file for data                               | _var_/system/users.json            |
| httpd.system.history                   | System file for data                               | _var_/system/history.json          |
| httpd.system.refresh                   | Controller information refresh interval            | 30s                                |
| httpd.system.windows.ok                | 'ok' time window after refresh                     | 10s                                |
| httpd.system.windows.uncertain         | 'uncertain' time window after last refresh         | 30s                                |
| httpd.system.windows.systime           | Allowed time window for controller system time     | 5m0s                               |
| httpd.system.windows.expires           | Cached controller attribute expiry time            | 2m0s                               |
| httpd.db.rules.acl                     | Grules file for fine-grained access control        | _etc_/httpd/acl.grl                |
| httpd.db.rules.interfaces              | Grules file for _interfaces_ admin authorisation   | _etc_/httpd/grules/interfaces.grl  |
| httpd.db.rules.controllers             | Grules file for _controllers_ admin authorisation  | _etc_/httpd/grules/controllers.grl |
| httpd.db.rules.cards                   | Grules file for _cards_ admin authorisation        | _etc_/httpd/grules/cards.grl       |
| httpd.db.rules.doors                   | Grules file for _doors_ admin authorisation        | _etc_/httpd/grules/doors.grl       |
| httpd.db.rules.groups                  | Grules file for _groups_ admin authorisation       | _etc_/httpd/grules/groups.grl      |
| httpd.db.rules.events                  | Grules file for _events_ admin authorisation       | _etc_/httpd/grules/events.grl      |
| httpd.db.rules.logs                    | Grules file for _logs_ admin authorisation         | _etc_/httpd/grules/logs.grl        |
| httpd.db.rules.users                   | Grules file for _users_ admin authorisation        | _etc_/httpd/grules/users.grl       |
| httpd.audit.file                       | Audit trail file                                   | _var_/httpd/audit/audit.log        |
| httpd.retention                        | Retention time for deleted items                   | 5m0s                               |
| httpd.timezones                        | File for custom timezones e.g. Afica/Cairo         | _etc_/timezones                    |


Sample HTTPD section:
```
# HTTPD
httpd.html = /usr/local/etc/com.github.uhppoted/http/html
httpd.http.enabled = true
; httpd.http.port = 8080
; httpd.https.enabled = true
; httpd.https.port = 8443
; httpd.tls.ca = /usr/local/etc/com.github.uhppoted/httpd/ca.cert
; httpd.tls.certificate = /usr/local/etc/com.github.uhppoted/httpd/uhppoted.cert
; httpd.tls.key = /usr/local/etc/com.github.uhppoted/httpd/uhppoted.key
httpd.tls.client.certificates.required = true
httpd.security.auth = some
; httpd.security.local.db = /usr/local/etc/com.github.uhppoted/httpd/auth.json
; httpd.security.cookie.max-age = 24
; httpd.security.login.expiry = 1m
httpd.security.session.expiry = 300s
httpd.request.timeout = 15s
; httpd.system.interfaces = /usr/local/var/com.github.uhppoted/httpd/system/interfaces.json
; httpd.system.controllers = /usr/local/var/com.github.uhppoted/httpd/system/controllers.json
; httpd.system.doors = /usr/local/var/com.github.uhppoted/httpd/system/doors.json
; httpd.system.groups = /usr/local/var/com.github.uhppoted/httpd/system/groups.json
; httpd.system.cards = /usr/local/var/com.github.uhppoted/httpd/system/cards.json
; httpd.system.events = /usr/local/var/com.github.uhppoted/httpd/system/events.json
; httpd.system.logs = /usr/local/var/com.github.uhppoted/httpd/system/logs.json
; httpd.system.users = /usr/local/var/com.github.uhppoted/httpd/system/users.json
; httpd.system.refresh = 30s
httpd.system.windows.ok = 10s
httpd.system.windows.uncertain = 30s
; httpd.system.windows.systime = 5m0s
; httpd.system.windows.expires = 2m0s
; httpd.db.rules.acl = /usr/local/etc/com.github.uhppoted/httpd/acl.grl
httpd.db.rules.interfaces = /usr/local/etc/com.github.uhppoted/httpd/grules/interfaces.grl
httpd.db.rules.controllers = /usr/local/etc/com.github.uhppoted/httpd/grules/controllers.grl
httpd.db.rules.cards = /usr/local/etc/com.github.uhppoted/httpd/grules/cards.grl
httpd.db.rules.doors = /usr/local/etc/com.github.uhppoted/httpd/grules/doors.grl
httpd.db.rules.groups = /usr/local/etc/com.github.uhppoted/httpd/grules/groups.grl
httpd.db.rules.events = /usr/local/etc/com.github.uhppoted/httpd/grules/events.grl
httpd.db.rules.logs = /usr/local/etc/com.github.uhppoted/httpd/grules/logs.grl
httpd.db.rules.users = /usr/local/etc/com.github.uhppoted/httpd/grules/users.grl
; httpd.audit.file = /usr/local/var/com.github.uhppoted/httpd/audit/audit.log
httpd.retention = 5m0s
; httpd.timezones = /usr/local/etc/com.github.uhppoted/timezones
```