# HTML

The HTML, CSS, images and Javascript files for the default implementation are [embedded](https://github.com/uhppoted/uhppoted-httpd/tree/master/httpd/html) in the executable and are loaded on startup.

As outlined in the [caveat emptor](https://github.com/uhppoted/uhppoted-httpd#raison-d%C3%AAtre), in line with the _uhppoted_
philosophy of providing components rather than solutions, the look and feel and implementation of the user interface has been
deliberately kept as simple and plain as it can reasonably be to simplify customising it on an individual basis. Specifically:

- the Javascript is plain vanilla Javascript (so no React or any other dependencies)
- vanilla and relatively uncomplicated CSS (generated from the _scss_ files in the 
  [Sass](https://github.com/uhppoted/uhppoted-httpd/tree/master/sass) folder

The HTML files do however use the _Go_ [templating engine](https://pkg.go.dev/html/template) to reuse common snippets of HTML -
the {{ ... }} markers in the HTML files indicate template replacements using the snippets found in the [templates](https://github.com/uhppoted/uhppoted-httpd/tree/master/httpd/html/templates) folder.


## Customising the user interface

### First steps

The static files for the base user interface can be obtained directly from the _github_ [repo](https://github.com/uhppoted/uhppoted-httpd)
or alternatively by running `uhppoted-httpd daemonize` and answering _yes_ when asked if you would like to unpack the HTML.

The first step in customising the user interface is to set `uhppoted-httpd` to use an external folder for the static
files:
```
uhppoted.conf

# HTTPD
httpd.html = /usr/local/./html
; httpd.http.enabled = true
; httpd.http.port = 8080
...
```

For development it may be simpler to enable HTTP by uncommenting the `httpd.http.enabled` and `http.http.port` lines in 
`uhppoted.conf`.

By default, the static files are loaded once on startup - running `uhppoted-httpd` in _debug_ mode reloads the files as necessary:
```
uhppoted-httpd --debug --console
```

### Folder structure

The default folder/file structure is as follows:

| *Folder/File*  | *Description*                                                                             |
| -------------- | ----------------------------------------------------------------------------------------- |
| css            | Folder containg CSS generated from Sass source files                                      |
| images         | PNGs, JPGs and SVGs                                                                       |
| fonts          | Font files                                                                                |
| javascript     | Javascript files                                                                          |
| sys            | HTML for each page in the user interface                                                  |
| templates      | Reusable snippets of common HTML                                                          |
| translations   | Basic support for languages other than English                                            |
| usr            | HTML files for _other_ pages                                                              |
| index.html     | The default HTML file used for non-specific page requests                                 |
| favicon.ico    | Firefox specifically requests /favicon.ico rather than the one in the HTML header         |
| manifest.json  |                                                                                           |

Permissions and routing is coded directly into the application so keeping to the above structure is recommended. If you
do need to modify it, the routing is located in:

- [httpd/httpd.go](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/httpd.go#L71-L104):
```
    mux.Handle("/css/", http.FileServer(fs))
    mux.Handle("/images/", http.FileServer(fs))
    mux.Handle("/javascript/", http.FileServer(fs))
    mux.Handle("/manifest.json", http.FileServer(fs))

    mux.HandleFunc("/sys/login.html", d.getNoAuth)
    ...
    mux.HandleFunc("/index.html", d.getNoAuth)
```

- [httpd/get.go](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/get.go#L35-L46):
```
    switch path {
    case "/interfaces",
        "/controllers",
        "/doors",
        "/cards",
        "/groups",
        "/events",
        "/logs",
        "/users":
        if handler := d.vtable(path); handler != nil && handler.get != nil {
            d.fetch(r, w, *handler)
        }
```

- [httpd/post.go](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/post.go#L21-L97):
```
    if path == "/authenticate" {
    ..

    if path == "/logout" {
    ..

    switch path {
    case "/password":
        ..

    case
        "/interfaces",
        "/controllers",
        "/doors",
        "/cards",
        "/groups",
        "/users":
        ..

    case "/synchronize/ACL":
        ..

    case "/synchronize/datetime":
        ..

    case "/synchronize/doors":
        ..
```

## API

The API is:
- relatively simple
- not a REST API
- probably going to change,

but for now comprises:

- [`get`](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/html/javascript/tabular.js#L567)
```
function get (urls, refreshed) {
...
}
```

takes a URL for a resource (e.g. /cards), retrieves a list of items for that resource as `{ OID, value }` pairs 
(see below), stores the information in the local _database_ and then invokes the `refreshed` function to process 
the update.

- [`post`](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/html/javascript/tabular.js#L704)
```
function post (url, created, updated, deleted, refreshed, reset, cleanup) {
...
}
```

takes a URL for a resource, along with:
   - list of objects to create
   - list of updated`{ OID, value }` pairs
   - list of `OIDs` to delete

and invokes the supplied `refreshed` function after the local database has been updated with the returned 
`{ OID, value}` pairs. The supplied `reset` functions is invoked to cleanup after a failed POST
request and the supplied `cleanup` function cleans up after both successful and failed POST requests.


### `{ OID, value }`

Every item on a page is tagged with an _Object Identication_ (OID) tag that uniquely identifies it to the 
backend. The [OID schema](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/OID.md) outlines
the structure of the OID tagspace (JS: [schema.js](https://github.com/uhppoted/uhppoted-httpd/blob/master/httpd/html/javascript/schema.js).

The OID approach was borrowed from SNMP and facilitates a flexible approach to populating information on a
page (GraphQL and REST were considered but typically require more structure than was deemed desirable - at
least at this point in time).
