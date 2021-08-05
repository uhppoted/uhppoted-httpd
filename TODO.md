## v0.7.x

### IN PROGRESS

- [ ] Make OID a type with:
      - HasPrefix
- [ ] Commonalise CSS

- Doors
  - [ ] 'global' OID cache
  - [ ] automatically create doors 1-4 for controllers
  - [ ] delete when name is blank and controller is not assigned
  - [ ] sort by controller + door
  - [ ] sort order different on Firefox and Chrome (sigh!)
  - [ ] save
  - [ ] Fix ACL for reworked types.Door
  - [ ] Deduplicate `type object status`
  - [ ] Move `doors.update` to tabular.js_
  - [ ] Move `doors.mark` to _tabular.js_
  - [ ] Move `doors.percolate` to _tabular.js_
  - [ ] Move `doors.set` to _tabular.js_
  - [ ] Move `onEdited`, `onRollback`, etc to _tabular.js_
  - [ ] Assign doors to controllers in _DOORS_ page and (optionally) make controller doors readonly

- [ ] Cards
      - migrate to OIDs

- [ ] system
      - configure ok/uncertain intervals
      - configure systime window
      - Fix `get-events` log string
      - Fix `get-status` log string
      - initialise LAN/all from uhppoted.conf (?)
      - logic around correcting time is weird
        -- enter to update doesn't always work
        -- set() is updating dataset.original which seems wrong but ...

      - add controller name to uhppote-core
      - add timezone to uhppote-core
      - validate Local::Device timezone on initialization
      - make DeviceID a type that handles nil on String() (like maybe Uint32 ???)
      - limit number of pending 'update' requests (e.g. if device is not responding)
      - use uhppoted-lib::healthcheck
      - move values to catalog

- [ ] Fix Firefox layout
      - internal table layout seems to include padding where Chrome doesn't
      - explicitly remove border from table rows/cells maybe (?)

- [ ] [TOML](https://toml.io) files

- [ ] tabular
      - New table row submitted with error cannot be discarded
      - Empty list: make first row a 'new' row
      - Commonalise controller and card handling into tabular.js

- [ ] ACL
      - wrap ACL update in goroutine
        -- error handling ??

- [ ] Loading bar a la cybercode
      - progresss
      - indefinite

- [ ] MemDB
      - Rather use sync.Map
      - JSON field names to lowercase
      - add created and modified timestamps to records
      - keep historical copies on save (for undo/revert)
      - default row order should be by 'created'
      - unit tests for ACL rules

- [ ] Card holders
      - Marshal cards.Records as "" if StatusUnknown
      - highlight current row (?)
      - unit tests for auth rules
      - card type should probably be a string (because otherwise 0 is a reserved number)
        -- 'nil' it if it's 0 ?
        -- think it through anyway
      - add
        - use cloneNode rather (https://stackoverflow.com/questions/1728284/create-clone-of-table-row-and-append-to-table-in-javascript)
        - shadow DOM ???
      - wrap templating in a decent error handler
        - redirect to error page

      - custom webelement (https://developer.mozilla.org/en-US/docs/Web/Web_Components/Using_custom_elements)
      - replace dataset.value with get()
      - draw out (TLA+ ?) local record FSM
        -- e.g. add -> delete on rollback deletes
        -- e.g. add -> commit only enabled after modified
      - genericize JS:refresh

      - filter columns
        -- pins!
      - undo/revert (?)
      - use internal DB rather than element dataset (?)
      - virtual DOM
      - search & pin
      - labels from translations
      - apply to all (pinned ?) columns
      - simultaneous editing (?) 
        -- use hash of DB to identify changes
        -- CRDT ??
      
- [ ] Events
      - https://jvns.ca/blog/2021/01/12/day-36--server-sent-events-are-cool--and-a-fun-bug/
      
- [ ] Login
      - include login cookie when redirecting to login.html to avoid the initial double click
      - restyle avatar to have a border and be a bit floaty (i.e. not be glued to top-right)

- [ ] favicon:https://nedbatchelder.com/blog/202012/favicons_with_imagemagick.html
- [ ] Use 'modular' naming convention for colours, etc. e.g. tabular-row-colour

- [ ] Fonts
- [ ] User settings
      - automatic logout enabled/timeout
      - change password
- [ ] Logo 
      - https://math.stackexchange.com/questions/3742825/why-is-the-penrose-triangle-impossible
      - https://jn3008.tumblr.com/post/618100274778783744
- [ ] Lighthouse test (Chrome dev tools)
- [ ] Hamburger menu (?)
- [ ] Style SVG icons with SASS
- [ ] Structure CSS somehow :-(
- [ ] Thoroughly rethink the whole timezone thing
- [ ] SCRAM authentication https://tools.ietf.org/html/rfc5802)
      - [SubtleCrypto](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto)
      - [PAKE](https://en.wikipedia.org/wiki/Password-authenticated_key_agreement) (?)

- [x] Deduplicate `status`
- [x] /favicon.ico
- [x] Signout page "doesn't have permission"
- [x] Neatened up login error reporting
- [x] Menu pops up when just vaguely over area
- [x] Make a nice synthesized HTML page for logout when server is down
- [x] Logout always i.e. ignore POST http://127.0.0.1:8080/logout net::ERR_CONNECTION_REFUSED
- [x] Logs out on POST to e.g. /system when not authorised. Should show error instead.
- [x] Warning message doesn't align left if commitall and/or rollbackall are not visible
- [x] 127.0.0.1/:323 The specified value "\u003Cnil\u003E" cannot be parsed, or is out of range.
- [x] Navigation
- [x] Not redirecting to login.html after restart
- [x] Fix 'login unauthorized'
- [x] Abstract authentication/authorization
- [x] Javascript modules
- [x] eslint
- [x] Fix 'message' bar 
- [x] Salt stored password hashes
- [x] Double GET index.html on login (?)
- [x] Login token
- [x] TLS
- [x] Automatic logout
- [x] Implement session
- [x] Show logged in user
- [x] Sign out
- [x] login page
- [x] SASS/CSS
- [x] Templatize HTML and set label text etc from file

## TODO

- [ ] [CRDT](https://concordant.io/software)
- [ ] [XHTML](https://www.nayuki.io/page/practical-guide-to-xhtml)
- [ ] Redesign using RDF/OWL triples ? 
      - https://github.com/severin-lemaignan/minimalkb
      - https://www.w3.org/TR/rdf11-primer/#section-triple
- [ ] 'Macro' keys
- [ ] Zootier input fields (e.g. https://css-tricks.com/float-labels-css)
- [ ] [Gradient borders](https://css-tricks.com/gradient-borders-in-css/)
- [ ] [JWK](https://tools.ietf.org/html/rfc7517)
- [ ] Support alternative auth providers e.g. auth0
- [ ] gitdb (?)
- [ ] [CRDT's](https://josephg.com/blog/crdts-are-the-future)
- [ ] UI widgets and frameworks:
      - [Shoelace](https://shoelace.style)
      - [WebFlow](https://www.toptal.com/designers/webflow/webflow-advantages)
      - [ExpertX](https://www.toptal.com/designers/webflow/webflow-advantages)
      - [gridstack](https://gridstackjs.com)
      - [toptal](https://www.toptal.com/designers/ux/notification-design)
      - [Tabulator](http://tabulator.info)
      - [Arwes](https://arwes.dev)
      - https://blog.datawrapper.de/beautifulcolors/
      - http://csszengarden.com/219
      - Colorways (for themes)
      - https://thenounproject.com
      - [retool](https://retool.com)
      - [plurid](https://github.com/plurid/plurid)


## NOTES

- [SVG favicon](https://medium.com/swlh/are-you-using-svg-favicons-yet-a-guide-for-modern-browsers-836a6aace3df)
- https://security.stackexchange.com/questions/180357/store-auth-token-in-cookie-or-header
- https://auth0.com/docs/tokens/concepts/token-storage
- https://stackoverflow.com/questions/12130582/setting-cookies-with-net-http
- https://thewhitetulip.gitbooks.io/webapp-with-golang-anti-textbook/manuscript/4.0authentication.html
- https://jonathanmh.com/example-json-web-tokens-vanilla-javascript/
- https://golangcode.com/api-auth-with-jwt/
- https://github.com/cristalhq/jwt
- [CSS Tabs](https://codepen.io/axelaredz/pen/ipome)
- [WenAuthN](https://trustfoundry.net/passwords-are-dead-long-live-webauthn)
- [ZUI](https://zircleui.github.io/docs/examples/home.html)
- [plurid](https://github.com/plurid/plurid)