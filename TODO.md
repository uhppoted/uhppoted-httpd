## v0.7.x

- [ ] **PANIC**
```
2022/01/28 11:08:58 DEBUG GET  /doors
fatal error: concurrent map read and map write

goroutine 2399 [running]:
runtime.throw({0x171c322, 0x2fbff7f0})
  /usr/local/go/src/runtime/panic.go:1198 +0x71 fp=0xc0002e7480 sp=0xc0002e7450 pc=0x1034a51
runtime.mapaccess2(0xc0752da3afbff7f0, 0x2c61fae59bb, 0x1cb39e0)
  /usr/local/go/src/runtime/map.go:469 +0x205 fp=0xc0002e74c0 sp=0xc0002e7480 pc=0x100fb05
github.com/uhppoted/uhppoted-httpd/auth.(*Local).extant(0xc0002009a0, 0x1, {0x11, 0x28, 0x7f, 0x7a, 0x80, 0x6d, 0x11, 0xec, ...})
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/auth/local.go:616 +0x114 fp=0xc0002e7570 sp=0xc0002e74c0 pc=0x1533f14
github.com/uhppoted/uhppoted-httpd/auth.(*Local).Authenticated(0xc0002009a0, {0xc0001f8036, 0x1737bf0})
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/auth/local.go:385 +0x10e fp=0xc0002e7670 sp=0xc0002e7570 pc=0x153222e
github.com/uhppoted/uhppoted-httpd/httpd/auth.(*Basic).Authenticated(0xc000126d80, 0x1713f27)
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/auth/basic.go:81 +0x49 fp=0xc0002e7778 sp=0xc0002e7670 pc=0x156cbc9
github.com/uhppoted/uhppoted-httpd/httpd.(*dispatcher).authenticated(0xc0002bdbc0, 0x1209e07, {0x17dc5b0, 0xc00018a000})
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/httpd.go:228 +0xeb fp=0xc0002e77f0 sp=0xc0002e7778 pc=0x15bbdab
github.com/uhppoted/uhppoted-httpd/httpd.(*dispatcher).fetch(0xc0002bdbc0, 0xc00040c100, {0x17dc5b0, 0xc00018a000}, {0x17375d0, 0x1737608})
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/get.go:206 +0xd5 fp=0xc0002e7950 sp=0xc0002e77f0 pc=0x15b90f5
github.com/uhppoted/uhppoted-httpd/httpd.(*dispatcher).get(0xc000454740, {0x17dc5b0, 0xc00018a000}, 0xc00040c100)
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/get.go:44 +0x1ca fp=0xc0002e7998 sp=0xc0002e7950 pc=0x15b7a6a
github.com/uhppoted/uhppoted-httpd/httpd.(*dispatcher).dispatch(0xc0002e7a30, {0x17dc5b0, 0xc00018a000}, 0xc00040c100)
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/httpd.go:213 +0x292 fp=0xc0002e7a18 sp=0xc0002e7998 pc=0x15bbbf2
github.com/uhppoted/uhppoted-httpd/httpd.(*dispatcher).dispatch-fm({0x17dc5b0, 0xc00018a000}, 0x0)
  /Users/tonyseebregts/Development/uhppote/uhppoted/uhppoted-httpd/httpd/httpd.go:202 +0x3c fp=0xc0002e7a48 sp=0xc0002e7a18 pc=0x15bec3c
net/http.HandlerFunc.ServeHTTP(0x0, {0x17dc5b0, 0xc00018a000}, 0x0)
  /usr/local/go/src/net/http/server.go:2046 +0x2f fp=0xc0002e7a70 sp=0xc0002e7a48 pc=0x12e196f
net/http.(*ServeMux).ServeHTTP(0x0, {0x17dc5b0, 0xc00018a000}, 0xc00040c100)
  /usr/local/go/src/net/http/server.go:2424 +0x149 fp=0xc0002e7ac0 sp=0xc0002e7a70 pc=0x12e3269
net/http.serverHandler.ServeHTTP({0xc000948f90}, {0x17dc5b0, 0xc00018a000}, 0xc00040c100)
  /usr/local/go/src/net/http/server.go:2878 +0x43b fp=0xc0002e7b80 sp=0xc0002e7ac0 pc=0x12e4edb
net/http.(*conn).serve(0xc0001b8460, {0x17e05e0, 0xc00011b290})
  /usr/local/go/src/net/http/server.go:1929 +0xb08 fp=0xc0002e7fb8 sp=0xc0002e7b80 pc=0x12e0a48
net/http.(*Server).Serve·dwrap·82()
  /usr/local/go/src/net/http/server.go:3033 +0x2e fp=0xc0002e7fe0 sp=0xc0002e7fb8 pc=0x12e582e
runtime.goexit()
  /usr/local/go/src/runtime/asm_amd64.s:1581 +0x1 fp=0xc0002e7fe8 sp=0xc0002e7fe0 pc=0x1064fe1
created by net/http.(*Server).Serve
  /usr/local/go/src/net/http/server.go:3033 +0x4e8

```

- [ ] Finish structuring catalog.schema
      - [ ] (?) Generate schema.js from catalog.Schema

- [ ] Clean up the log function mess in controller

- [ ] Fix DateTime mess
      - [ ] Combobox for datetime
            - [ ] Remove unused aria stuff
            - [ ] Style for dark mode
            - [ ] Initialise options with suggested values
                   - (?) datalist (https://demo.agektmr.com/datalist)
                   - https://devhints.io/wip/intl-datetime
                   - https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/DateTimeFormat/DateTimeFormat
            - [ ] Make timezones a map between displayed and actual e.g. GMT+2 -> Etc/GMT+2
            - [ ] Fix keydown/keyup from input field
            - [ ] Fix list style
            - [x] Move combobox.css to sass
            - [x] Let list overflow table (https://css-tricks.com/popping-hidden-overflow)
            - [x] Remove button
            - [x] Rename combobox.js to datetime.js
            - (?) datetime-local
            - (?) Use shadow DOM

      - [ ] (?) Treat empty DateTime as the null value
      - [ ] (?) Change Format(), etc to not take pointer receiver
      - [ ] (?) Define FormatDateTime() method that *does* take pointer
      - [ ] (?) Refactor DateTime out to use core implementation only
      - [ ] (?) Implement Ptr() for the repetitive assign to variable to take address thing
      - (thoroughly relook at the whole timezone thing)
      - [x] Let timezone(..) just return Local instead of getting all complicated about it
      - [x] Do the 'uncertain/pending' thing for timezones
      - [x] Add `datetime.modified` to controller.cached

### IN PROGRESS

- [ ] OIDs:
      - [ ] catalog.Objects type to streamline e.g. append, trim, etc
      - [ ] GetV => GetBool, GetInt, etc
      - [ ] MAYBE: Store all values in catalog and 'realize' local copies from cache

- [ ] 'users' page
- [ ] daemonize/undaemonize
      - [ ] set initial user + password
- [ ] Rename 'address' to 'endpoint'
      - https://networkengineering.stackexchange.com/questions/9429/what-is-the-proper-term-for-ipaddress-hostnameport

#### Doors
  - [ ] Custom 'mode' dropdown to handle option click so that list can be updated asynchronously
        - https://w3c.github.io/aria-practices/examples/combobox/combobox-select-only.html
        - https://stackoverflow.com/questions/3518002/how-can-i-set-the-default-value-for-an-html-select-element
  - [ ] 'door' select chooses first item if list changes while select is open

#### Events
  - [ ] Optimize page display
        - [ ] Realize e.g. two pages and repopulate OIDs a la Android RecyclerView
              - https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
        - [ ] Render/realize only if updated
        - [ ] (?) Keep DB in local storage

#### Cards
  - [ ] Weirdness around card add/delete
        - [ ] Return error for edits to card without name or number (e.g.'new' card)
        - [ ] What happens if other edits happen before card name/number is updated (e.g. for delete/add)?

  - [ ] Fix bottom right of scrollbar
        - [ ] Scrollbar 'goes funny' if -webkit styles are modified

  - [ ] Rethink CardHolder.Card (pointer implementation is unnecessarily messy)
  - [ ] `refresh` is overwriting pending group edits
  - [ ] Replace dataset.original with value from DB
  - [ ] Unit test for AsObjects

#### System
      - [ ] replace audit.module value with something more usefully loggable e.g. C:deviceID:name
      - [ ] (?) Update interfaces.js to defer to tabular.js
            - [ ] rollback
            - [ ] commit
            - [ ] set

      - [ ] Rethink controller device ID (pointer implementation is unnecessarily messy)
            - (maybeeeeeeee) make generic type for uint32 that handles nil/0 on String()
            - 'natural' number type :-)

      - Export to uhppoted.conf
        - 'export' command line argument 
        - 'export' admin menu option
        - 'auto-export' option (?)

      - (?) Import from uhppoted.conf
        - 'import' command line argument 
        - 'import' admin menu option
        - 'auto-import' option (?)

      - logic around correcting time is weird
        -- enter to update doesn't always work
        -- set() is updating dataset.original which seems wrong but ...

      - add controller name to uhppote-core
      - add timezone to uhppote-core
      - validate Local::Device timezone on initialization
      - limit number of pending 'update' requests (e.g. if device is not responding)
      - use uhppoted-lib::healthcheck

#### Other
- [ ] Fix Firefox layout
      - spacing/padding/margins

- [ ] [TOML](https://toml.io) files

- [ ] tabular
      - (experiment) use :before or :content for flags???
      - New table row submitted with error cannot be discarded
      - Empty list: make first row a 'new' row (?)
      - filter columns
      - genericize JS:refresh
      - 'faceted' filtering (https://ux.stackexchange.com/questions/48992/sorting-filtering-on-infinite-scrolling-data-table)

- [ ] ACL
      - wrap ACL update in goroutine
        -- error handling ??

- [ ] Loading bar a la cybercode
      - progresss
      - indefinite

- [ ] MemDB
      - Rather use sync.Map
      - keep historical copies on save (for undo/revert)
        - git (?)
      - unit tests for ACL rules

- [ ] Cards
      - unit tests for auth rules
      - card type should probably be a string (because otherwise 0 is a reserved number)
        -- 'nil' it if it's 0 ?
        -- think it through anyway
      - add
        - shadow DOM ???
      - wrap templating in a decent error handler
        - redirect to error page

      - custom webelement (https://developer.mozilla.org/en-US/docs/Web/Web_Components/Using_custom_elements)

      - pin selected rows
      - virtual DOM
      - search & pin
      - labels from translations
      - apply to all (pinned ?) columns
      - simultaneous editing (?) 
        -- use hash of DB to identify changes
      
- [ ] favicon:https://nedbatchelder.com/blog/202012/favicons_with_imagemagick.html
- [ ] Use 'modular' naming convention for colours, etc. e.g. tabular-row-colour

- [ ] Fonts
- [ ] User settings
      - automatic logout enabled/timeout
- [ ] Logo 
      - https://math.stackexchange.com/questions/3742825/why-is-the-penrose-triangle-impossible
      - https://jn3008.tumblr.com/post/618100274778783744

## TODO

- [ ] Hamburger menu (?)
- [ ] SCRAM authentication https://tools.ietf.org/html/rfc5802)
      - [SubtleCrypto](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto)
      - [PAKE](https://en.wikipedia.org/wiki/Password-authenticated_key_agreement) (?)

- [ ] Server events in addition to/rather-than refresh
      - https://jvns.ca/blog/2021/01/12/day-36--server-sent-events-are-cool--and-a-fun-bug/
- [ ] Lighthouse test (Chrome dev tools)
- [ ] [CRDT](https://concordant.io/software)
       - https://josephg.com/blog/crdts-go-brrr
       - [Braid](https://braid.org
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
      - []

- Model editing in TLA+
  -- e.g. add -> delete on rollback deletes
  -- e.g. add -> commit only enabled after modified

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
- [git/content-addressable filesystem](https://jvns.ca/blog/confusing-explanations)
- [Firefox: bug #1730211](https://bugzilla.mozilla.org/show_bug.cgi?id=1730211)
- https://stackoverflow.com/questions/1728284/create-clone-of-table-row-and-append-to-table-in-javascript
- https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
- https://stackoverflow.com/questions/11688279/jquery-infinite-scroll-on-a-table
- https://uxdesign.cc/build-an-infinite-scroll-table-without-scroll-event-listener-5949ce8e9a32
- https://www.dusanstam.com/posts/material-ui-table-with-infinite-scroll
- http://scrollmagic.io/examples/advanced/infinite_scrolling.html
- https://github.com/janpaepke/ScrollMagic
- [JSON-LS](https://json-ld.org)
- [Microdata](https://html.spec.whatwg.org/multipage/microdata.html#microdata)

# REFERENCES

- https://stackoverflow.com/questions/40328932/javascript-es6-promise-for-loop
