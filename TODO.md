## v0.7.x

- [ ] Make login cookie expire in 60s
      - Don't check login cookie except for login

- [ ] OIDs:
      - [ ] Clean up controller schem
            - [x] restructure doors
            - [ ] restructure address
            - [ ] restructure date/time
            - [ ] restructure cards
            - [ ] restructure events
      - [ ] Clean up door schema
            - [ ] Remove controller OID
            - [ ] Remove controller device ID
            - [x] Remove controller name
            - [x] Remove controller created
            - [ ] Remove controller door
      - [x] Use embedded struct for schema metadata

### IN PROGRESS

- [ ] Apply fix for 'interim hack':
      - [ ] Use deleted OID rather than status
            - [ ] controllers
            - [ ] cards
            - [ ] doors
            - [ ] groups
      - [ ] (?) Return status if IsValid && !IsDeleted
      - [ ] With the way db.js works now, returning a 'deleted' status will recreate an
            object that has been swept

- [ ] Genericize load/save
      - [ ] Define system.Component interface
      - [ ] Read file in system and pass in json.RawMessage to Load(...)
            - [ ] Move file to system
                  - [ ] groups
                  - [ ] interfaces
                  - [ ] controllers
                  - [ ] doors
                  - [ ] cards
            - [ ] Restructure to support e.g. DB as an alternative
      - [ ] Save file in system and get json.RawMessage from Save(...)
            - [ ] groups
            - [ ] cards
            - [ ] doors
            - [ ] controllers
            - [ ] interfaces
  
- [ ] 'users' page
- [ ] Set initial user + password
- [ ] Rename 'address' to 'endpoint'
      - https://networkengineering.stackexchange.com/questions/9429/what-is-the-proper-term-for-ipaddress-hostnameport

#### Doors
  - [ ] Custom 'mode' dropdown to handle option click so that list can be updated asynchronously
        - https://w3c.github.io/aria-practices/examples/combobox/combobox-select-only.html
        - https://stackoverflow.com/questions/3518002/how-can-i-set-the-default-value-for-an-html-select-element
  - [ ] 'door' select chooses first item if list changes while select is open

#### OID
  - [ ] GetV => GetBool, GetInt, etc
  - [ ] Store all values in catalog and 'realize' local copies from cache
  - [ ] Check RWLock for clone to make shadow
        - [ ] controllers
        - [ ] doors
        - [ ] cards
        - [ ] groups

#### Events
  - [ ] Replace LAN callback with something more idiomatic
  - [ ] Optimize page display
        - [ ] Realize e.g. two pages and repopulate OIDs a la Android RecyclerView
              - https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
        - [ ] Render/realize only if updated
        - [ ] (?) Keep DB in local storage

#### Cards
  - [ ] Weirdness around card add/delete
        - [ ] Handle edits to 'new' card that don't e.g. update the name or number
              - Return error (?)
              - Don't delete card

  - [ ] Fix bottom right of scrollbar
        - [ ] Scrollbar 'goes funny' if -webkit styles are modified

  - [ ] Rethink CardHolder.Card (pointer implementation is unnecessarily messy)
  - [ ] Commonalise load/save/print implementation
  - [ ] `refresh` is overwriting pending group edits
  - [ ] Replace dataset.original with value from DB
  - [ ] Unit test for AsObjects
  - [ ] Rethink mark/sweep to not use a counter
        - mark/sweep can be called multiple times for the same update
        - time based (?)
        - make update ID base (?)
        - use tag (?)

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
      - GET /system, /doors , /cards, etc all return everything. Need finer grained access 
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

