# Roadmap

## Design Notes

Although `uhppoted-httpd` will (eventually) be a fully functional user interface for managing UHPPOTE 
TCP/IP controllers, the design and implementation are intended to fit into the 'set of components' 
philosophy that backs the other`uhppoted` modules i.e. it is should to fit into the working systems
of the user and/or organisation, rather than have an identity of it's very own.

As such, the UI is intentionally simple, plain, low key and relatively unopinionated. It is intended 
to be a working UI that can be customised with relatively little effort (logo's, themes and CSS). 
Likewise, the scripting is vanilla Javascript (rather than e.g. Typescript or React) to keep the 
complexity of the system to a reasonable level - which should hopefully also facilitate low maintenance
in the long term.

## Road map

The list below is a provisional list of features and functionality that are on the road map:

#### v0.8.0

v0.8.0 is intended to provide the base layer functionality for a UI that manages 'local' controllers
(i.e. directly accessible via the network), backed by an in-memory database. Provisionally, the 
supported functionality will include:

- User ID+password authentication and authorization
- HTTP and HTTPS support
- Table based card management
- Table based controller management
- Door access rules
- Events view
- Logs view (?) 
- Websocket protocol for real-time'ish events and controller statuses
- Switchable UI themes

#### v0.8.1

v0.8.1 is (provisionally) envisioned as adding UI support for `uhppoted-rest` as well as an optional SQL backend 
database.

- Add support for controllers accessible via uhppoted-rest
- Optional SQLite database
- Greasemonkey/Tampermonkey support

#### v0.8.2

v0.8.2 is (provisionally) envisioned as adding UI support for `uhppoted-mqtt`.

- Add support for controllers accessible via uhppoted-mqtt
- Add support for NoSQL backend database (?)
- OAuth authentication

#### v0.8.3

v0.8.3 is (provisionally) envisioned as adding UI support for controller accessed via the file based
`uhppoted-app-s3`.

- Add support for controllers accessible via `uhppoted-app-s3`

#### vX.X.X

Far future features:

- UQL
- Query UI
- CRDT 
- Multi-tenant support
