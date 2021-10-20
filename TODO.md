## v0.7.x

- [ ] Rename 'address' to 'endpoint'
      - https://networkengineering.stackexchange.com/questions/9429/what-is-the-proper-term-for-ipaddress-hostnameport

### IN PROGRESS

- [ ] Server events in addition to/rather-than refresh
      - https://jvns.ca/blog/2021/01/12/day-36--server-sent-events-are-cool--and-a-fun-bug/
- [ ] Show 'offline' status on 'NET:CONNECTION REFUSED/TypeError: failed to fetch' but still logged in
- [ ] Double click for login after 'idle signout'
- [ ] 'reload' crashes with after restarting httpd
      ```
      tabular.js:416 TypeError: Cannot read properties of null (reading 'dataset')
          at unbusy (uhppoted.js:176)
          at tabular.js:394
      ```
- [ ] 'reload' alert 'undefined' message on restarting httpd
- [ ] 'reload' automatically if httpd comes alive again

#### auth
  - [ ] Reload grules file if changed
  - [ ] Pass UID/role to grule

#### OID
  - [ ] Store all values in catalog and 'realize' local copies from cache
        - Hmmm, may cause issues with shadow logic
        - Unless use shadow cache ??
        - .. or global mutex :-(
        - only ever used for doors so maybe just use catalog as a pointer to actual thing?

  - [ ] Only update catalog values after validate i.e. not in set(...)
        - [ ] rethink the whole 'dirty' thing
              - maybe use a stash queue (?)
              - or callbacks a la events
              - not callbacks -> channels!!!
        - [ ] Check RWLock for clone
              - [ ] controllers
              - [ ] doors
              - [ ] cards
              - [ ] groups

  - [ ] Make OID a type
        - [ ] move all _stringify's_ to OID/object/somesuch

#### Events
  - [ ] (?) Rework to use channels
  - [ ] (?) Genericize load/save for migration to MemDB
  - [ ] Optimize page display
        - [ ] Realize e.g. two pages and repopulate OIDs
              - https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
        - [ ] Render/realize only if updated
        - [ ] (?) Keep DB in local storage
  - [ ] Lookup historical card/door/controller assignments

#### Cards
  - [ ] Update unit tests for OID'd implementation
        - Make audit trail usable for unit tests
  - [ ] Weirdness around adding card
        - At top of list until updated
        - Can't delete card with name but no number
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

#### Groups

#### Logs
  - [ ] Format log records for 'add'
        - [x] doors
        - [x] groups
        - [x] cards
  - [ ] Format log records for 'delete'
        - [x] doors
        - [x] groups
        - [x] cards

  - [ ] Replace audit log record `Info` with details field
  - [ ] Commit audit trail after validate + save
  - [ ] Save
  - [ ] Load

#### Doors
  - [ ] Updates all _incorrect_ values if one item is edited
  - [ ] Custom 'mode' dropdown to handle option click so that list can be updated asynchronously
        - https://w3c.github.io/aria-practices/examples/combobox/combobox-select-only.html
        - https://stackoverflow.com/questions/3518002/how-can-i-set-the-default-value-for-an-html-select-element
  - [ ] 'door' select chooses first item if list changes while select is open

#### System
      - [ ] replace audit.module value with something more usefully loggable e.g. C:deviceID:name
      - [ ] (?) Update interfaces.js to defer to tabular.js
            - [ ] rollback
            - [ ] commit
            - [ ] set

      - [ ] Rethink controller device ID (pointer implementation is unnecessarily messy)
            - (maybeeeeeeee) make generic type for uint32 that handles nil/0 on String()
      - [ ] Fix `get-events` log string
      - [ ] Fix `get-status` log string
      - Input + datalist for timezone ?????
        - https://demo.agektmr.com/datalist/
      - Export to uhppoted.conf
        - 'export' command line argument 
        - 'export' admin menu option
        - 'auto-export' option (?)

      - Import from uhppoted.conf
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
      - JSON field names to lowercase
      - add created and modified timestamps to records
      - keep historical copies on save (for undo/revert)
      - default row order should be by 'created'
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
      - draw out (TLA+ ?) local record FSM
        -- e.g. add -> delete on rollback deletes
        -- e.g. add -> commit only enabled after modified

      - pin selected rows
      - virtual DOM
      - search & pin
      - labels from translations
      - apply to all (pinned ?) columns
      - simultaneous editing (?) 
        -- use hash of DB to identify changes
        -- CRDT ??
      
- [ ] Security
      - GET /system, /doors , /cards, etc all return everything. Need finer grained access 
      - Templates have access to everything - need finer grained access 

- [ ] Login
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
- [ ] Thoroughly rethink the whole timezone thing
- [ ] SCRAM authentication https://tools.ietf.org/html/rfc5802)
      - [SubtleCrypto](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto)
      - [PAKE](https://en.wikipedia.org/wiki/Password-authenticated_key_agreement) (?)

## TODO

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

# REFERENCES

- https://stackoverflow.com/questions/40328932/javascript-es6-promise-for-loop

