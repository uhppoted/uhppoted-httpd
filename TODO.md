# TODO

### IN PROGRESS

- [ ] https://github.com/uhppoted/uhppoted-httpd/issues/13
      - [ ] _admin_ unlock user
            - [ ] Refresh not working after unlock
                  - (probably not after reset OTP either)
            - [x] Need to reset failed login count after unlock
            - [x] (maybe) move all the locking stuff to system.User ???
            - [x] unlock log message: `user33%!(EXTRA string=Mike Mouz)`
            - [ ] What about the Local users map???
      - [ ] Make User.OTPKey the same as User.Password i.e. not _gettable_

- [x] Crash on Ok with empty passwords
      - [ ] Check auth.IsNil for all xxx.ToObjects(auth)

- [ ] [OKSolar](https://meat.io/oksolar)
- [ ] Set cookie.Secure to true for TLS requests
- [ ] Top border radius on overview events and log panels
- [ ] Figure out some way to secure users.json against manual rewriting

#### Doors
  - [ ] Custom 'mode' dropdown to handle option click so that list can be updated asynchronously
        - https://stackoverflow.com/questions/3518002/how-can-i-set-the-default-value-for-an-html-select-element
  - [ ] 'door' select chooses first item if list changes while select is open

#### Events
  - [ ] Optimize page display
        - [ ] Realize e.g. two pages and repopulate OIDs a la Android RecyclerView
              - https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
        - [ ] Render/realize only if updated
        - [ ] (?) Keep DB in local storage

#### Cards
  - [ ] `refresh` is overwriting pending group edits
  - [ ] Replace dataset.original with value from DB

#### System
      - validate Local::Device timezone on initialization
      - limit number of pending 'update' requests (e.g. if device is not responding)

## FYI

- [ ] Look into [Temporal](https://blogs.igalia.com/compilers/2020/06/23/dates-and-times-in-javascript)
      for date/time stuff
- [ ] https://tls-anvil.com/docs/Quick-Start/index
- [ ] https://jakub-m.github.io/2022/07/17/laport-clocks-formal.html
- (?) [UCAN](https://ucan.xyz/)
- [ ] System font stack (https://llccing.github.io/30-seconds-of-css/)

## TODO

- (?)s Rework login to use Authorization header with Basic/Digest
- [ ] Use browser local storage for DB
      - (?) ETags
            - https://ieftimov.com/posts/conditional-http-get-fastest-requests-need-no-response-body/

- https://www.youtube.com/watch?v=24GRiOCa1Vo
- https://www.theregister.com/2022/06/20/redbean_2_a_singlefile_web
- https://github.com/letoram/pipeworld

- (?) Multi-tenant
      - https://stanislas.blog/2021/08/firecracker/

- (?) Appliance
      - https://gokrazy.org
- (?) Rearchitecture as data+rules
- (?) Rearchitecture with channels 
- [ ] Rethink passing DBC to every call - it's only for the logs and maybe the audit trail could
      be updated from the catalog ??
      - (?) broadcast channel
      - (?) event bus
      - (?) condition handlers a la Lisp
      - (?) package audit logger
      - ... although ... could be useful for (upcoming) server sent events

- (?) 'Natural' number type (i.e. starts at 1) for device ID, card number, etc

- [ ] Fix Firefox layout
      - https://css-tricks.com/snippets/css/better-helvetica/

- [ ] Loading bar a la cybercode
      - progresss
      - indefinite

- [ ] User menu
      - automatic logout enabled/timeout
      - (?) Export to uhppoted.conf
            - 'export' command line argument 
            - 'export' admin menu option
            - 'auto-export' option (?)

      - (?) Import from uhppoted.conf
            - 'import' command line argument 
            - 'import' admin menu option
            - 'auto-import' option (?)

### Cleanup

- [ ] MAYBE: Store all values in catalog and 'realize' local copies from cache
- [ ] Rename 'address' to 'endpoint'
      - https://networkengineering.stackexchange.com/questions/9429/what-is-the-proper-term-for-ipaddress-hostnameport
- [ ] (?) Generate schema.js from catalog.Schema
- [ ] favicon:https://nedbatchelder.com/blog/202012/favicons_with_imagemagick.html
- [ ] Logo 
      - https://math.stackexchange.com/questions/3742825/why-is-the-penrose-triangle-impossible
      - https://jn3008.tumblr.com/post/618100274778783744
- [ ] Use 'modular' naming convention for colours, etc. e.g. tabular-row-colour

- [ ] Cards
      - unit tests for auth rules
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
      
- [ ] MemDB
      - Rather use sync.Map
      - keep historical copies on save (for undo/revert)
        - git (?)
      - unit tests for ACL rules

- [ ] tabular
      - Empty list: make first row a 'new' row (?)
      - filter columns
      - genericize JS:refresh
      - 'faceted' filtering (https://ux.stackexchange.com/questions/48992/sorting-filtering-on-infinite-scrolling-data-table)


### Functionality

- (?) Restructure using event sourcing
  -- https://kickstarter.engineering/event-sourcing-made-simple-4a2625113224
  -- [CQRS](https://docs.microsoft.com/en-us/azure/architecture/patterns/cqrs)
  

- (?) Morton codes for catalog

- [ ] [TOML](https://toml.io) files
- [ ] Hamburger menu (?)
- [ ] Use shadow DOM for datetime combobox
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
- [Vercel ](https://vercel.com)
- [Riffle](https://riffle.systems/essays/prelude)
- [SASL](https://inspektor.cloud/blog/password-based-authentication-without-tls-using-sasl)
  - https://en.wikipedia.org/wiki/Salted_Challenge_Response_Authentication_Mechanism
- https://maori.geek.nz/golang-desktop-app-webview-vs-lorca-vs-electron-a5e6b2869391
- https://github.com/wailsapp/wails

# REFERENCES

- https://stackoverflow.com/questions/40328932/javascript-es6-promise-for-loop
- [datalist](https://demo.agektmr.com/datalist)
- https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/DateTimeFormat/DateTimeFormat
- https://gist.github.com/Mattemagikern/328cdd650be33bc33105e26db88e487d

