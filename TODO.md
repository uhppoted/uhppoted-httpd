## v0.7.x

- [ ] Update events handling
      - [ ] Rethink not sending 0 for events.first, events.last and events.current
            - (?) Maybe make Uint32 i.e. 0 is ""
      - [ ] Remove LAN.store

- [ ] Restyle highlighted fields in _dark_ mode (e.g. after editing controller name)

- [ ] Clean up HTTP server
      - [x] Move static file handling out of httpd.get
      - [x] Separate handler for GET login/unauthorised/etc
      - [x] Separate handler for GET *.html
      - [ ] (?) Separate GET+POST+HEAD dispatcher for /<data>
      - [ ] (?) Use ServeMux (for all the sanitization that comes with it)
      - [ ] Update httpdFileSystem to use FS 
      - [ ] Unit test httpdFileSystem for
            - dot file hiding
            - path escaping
      - [ ] Remove auth.Basic session `sweep`
            - [x] Invalidate session ID in auth.Local on logout
            - [x] Remove session stuff from auth.Basic
            - [ ] Keep session ID internal to auth.Local
            - [ ] Make auth.Local constants internal

      - [x] Regenerate session keys every N minutes
            - [x] Replace fixed serialized key with generated key
            - [x] Regenerate session key
            - [x] Replace single key with key list
            - [x] Verify/Authenticated with every key in list
            - [x] Refresh session cookie when necessary
            - [x] Spurious _jwt: key is nil_ warnig because SessionCookie not valid after a restart
      - [x] Rework auth.Local.claims to use JSON encoded login and session fields
      - [x] Clean up GET
      - [x] Clean up auth.Local
      - [x] auth.Basic.Authenticated redundantly checks token in both authenticated and session
      - [x] Lift `unauthorised` handling out of auth provider
      - [x] Make default login cookie expiration 60s
      - [x] Clear login cookie on login
      - [x] Clear session cookie on logout
      - [x] Clear session cookie on unauthenticated
      - [x] GET allow unauthenticated/authorized access to CSS, images, etc
      - [x] GET allow unauthenticated access to login.html and unauthorized.html
      - [x] GET require authenticated + authorisation for access to data
      - [x] GET require authenticated + authorisation for access to everything else
      - [x] GET return 'Not Found' for arbitrary stuff
      - [x] HEAD restrict to /authenticate
      - [x] POST allow unauthenticated access to /authenticate
      - [x] POST allow unauthenticated access to /logout because httpd may have restarted, invalidating sessions
      - [x] POST require authenticated+authorisation for access to everything else
      - [x] POST require login cookie for /authenticate
      - [x] Commonalize httpd.unauthenticated()
      - [x] Commonalize httpd.unauthorized()
      - [x] Remove authorised resource check in basic.Verify (muddled responsibility)
      - [x] Remove authorised resource check in basic.SetPassword (muddled responsibility)
      - [x] Clean up auth.Basic
      - [x] Touch session when authenticated
      - [x] Prevent e.g. CSS/../events.html from poking a hole in the auth framework
            - [x] GET
            - [x] HEAD
            - [x] POST
            - [x] Auth provider should take login cookie not request
            - [x] Auth provider should take session cookie not request
            - [x] Use resolved path in auth provider i.e. don't pass http.Request to auth provider
                  - [x] Authenticated
                  - [x] Logout
                  - [x] Session
      - [x] Homogenize `authorised` and `authorisedX`
      - [x] Don't check login cookie except for login

- [ ] Fix DateTime mess
      - [ ] Controller schema for DateTime
      - [ ] Do the 'uncertain/pending' thing for timezones
      - [ ] MAYBE: treat empty DateTime as the null value
      - [ ] MAYBE: change Format(), etc to not take pointer receiver
      - [ ] MAYBE: define FormatDateTime() method that *does* take pointer
      - [ ] Refactor DateTime out to use core implementation only
      - [ ] Implement Ptr() for the repetitive assign to variable to take address thing

### IN PROGRESS

- [ ] OIDs:
      - [ ] catalog.Objects type to streamline e.g. append, trim, etc
      - [ ] GetV => GetBool, GetInt, etc
      - [ ] MAYBE: Store all values in catalog and 'realize' local copies from cache

- [ ] 'users' page
- [ ] Set initial user + password
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
      - Input + datalist for timezone ?????
        - https://demo.agektmr.com/datalist/

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
- [ ] (?) _heartbeat_ for online/offline

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
      
- [ ] Security
      - Templates have access to everything - need finer grained access 

- [ ] favicon:https://nedbatchelder.com/blog/202012/favicons_with_imagemagick.html
- [ ] Use 'modular' naming convention for colours, etc. e.g. tabular-row-colour

- [ ] Fonts
- [ ] User settings
      - automatic logout enabled/timeout
      - change password
- [ ] Logo 
      - https://math.stackexchange.com/questions/3742825/why-is-the-penrose-triangle-impossible
      - https://jn3008.tumblr.com/post/618100274778783744
- [ ] Hamburger menu (?)
- [ ] Thoroughly rethink the whole timezone thing

## TODO

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
