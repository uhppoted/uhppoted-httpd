# `auth.json`

The `auth.json` file sets the coarse grained authorisation for user access to resource URLs for GET and POST requests. It is
variously located in the _uhppoted-httpd_ configuration folder:

- /etc/uhppoted/httpd/auth.json (Linux)
- /usr/local/com.github.uhppoted/httpd/auth.json (MacOS)
- \Program Data\uhppoted\httpd\auth.json (Windows)

A typical entry looks like:
```
{
  "resources": [
    {
      "path": "^/index.html$",
      "authorised": ".*"
    },
    {
      "path": "^/sys/controllers.html$",
      "authorised": "^(admin|user)$"
    },
    {
      "path": "^/sys/controllers$",
      "authorised": "^(admin)$"
    },
    ...
  ]
}
```

and comprises:

- the `path` to the resource URL, expressed as a regular expression
- the `authorised` roles, also expressed as a regular expression

In the sample above:

- `/index.html` is unrestricted i.e. viewable by anybody
- `/sys/controllers/html` is viewable by logged in users with either an _admin_ or a _user_ role
- `/sys/controllers` (the URL for making changes to the controllers) is only allowed for logged in users with an _admin_ role

User roles can be set/edited on the _users_ page in the user interface.

## Resources

The default list of resources comprises:

|Path                       | Method   | Description                                                      |
|---------------------------|----------|------------------------------------------------------------------|
| /index.html               | GET      | Default URL - redirects to either the _login_ or _overview_ page |
| /favicon.ico              | GET      | _favicon_ (for Firefox)                                          |
| /sys/login.html           | GET      | Login page                                                       |
| /sys/unauthorized.html    | GET      | Redirect page for unauthorised requests                          |
| /sys/overview.html        | GET      | System summary page for logged in users                          |
| /sys/controllers.html     | GET      | Controller details page                                          |
| /sys/cards.html           | GET      | Card details page                                                |
| /sys/doors.html           | GET      | Access controlled doors details page                             |
| /sys/groups.html          | GET      | Access control groups details page                               |
| /sys/events.html          | GET      | Access control events list                                       |
| /sys/logs.html            | GET      | Access control log records list                                  |
| /sys/users.html           | GET      | User name,password and role adminstration page                   |
| /sys/password.html        | GET      | User password maintenance page                                   |
| /other.html               | GET      | Place holder for 'other' pages                                   |
| /password                 | GET/POST | Update password POST requests                                    |
| /interfaces               | GET/POST | View/create/update/delete interface configuration                |
| /controllers              | GET/POST | View/create/update/delete controller configuration               |
| /doors                    | GET/POST | View/create/update/delete door configuration                     |
| /cards                    | GET/POST | View/create/update/delete card information                       |
| /groups                   | GET/POST | View/create/update/delete access control groups                  |
| /events                   | GET      | Retrieves access control events                                  |
| /logs                     | GET      | Retrieves access control log records                             | 
| /users                    | GET/POST | View/create/update/delete user records                           |
| /synchronize/ACL          | POST     | Synchronize access control list across all controllers           |
| /synchronize/datetime     | POST     | Synchronize date/time across all controllers                     |
| /synchronize/doors        | POST     | Synchronize door configuration across all controllers            |
| /otp                      | POST     | Create and revoke user OTPs                                      |

