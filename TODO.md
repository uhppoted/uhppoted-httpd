## v0.7.x

### IN PROGRESS

- [ ] Loading bar a la cybercode

- [ ] ACL
      - groups => rules => permissions
      - wrap ACL update in goroutine
        -- error handling ??

- [ ] MemDB
      - JSON field names to lowercase
      - add created and modified timestamps to records
      - keep historical copies on save (for revert)
      - default row order should be by 'created'

- [ ] Card holders
      - add/update/delete authorisation
        - initialise from config
        - unit tests
      - commit multiple rows
        - commit-all
        - rollback-all
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
      - draw out local record FSM
        -- add -> rollback deletes, commit only enabled after modified
      - genericize JS:refresh
      - rename Go:Permissions to e.g. Membership

      - undo/revert (?)
      - use internal DB rather than element dataset (?)
      - scroll horizontally
      - scroll vertically
      - virtual DOM
      - search & pin
      - gzip response
      - labels from translations
      - scroll horizontally
      - scroll vertically
      - freeze header rows and columns
      - filter columns
        -- pins!
      - apply to all (pinned ?) columns
      - simultaneous editing (?) 
        -- use hash of DB to identify changes
        -- CRDT ??
      
- [ ] Login
      - include login cookie when redirecting to login.html to avoid the initial double click
      - restyle avatar to have a border and be a bit floaty (i.e. not be glued to top-right)

- [ ] Take a look at:
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

- [ ] Fonts
- [ ] favicon
      - convert text to paths and cleanup SVG
      - Safari
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
- [ ] SCRAM authentication https://tools.ietf.org/html/rfc5802)
      - [SubtleCrypto](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto)
      - [PAKE](https://en.wikipedia.org/wiki/Password-authenticated_key_agreement) (?)

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

- [ ] Zootier input fields (e.g. https://css-tricks.com/float-labels-css)
- [ ] [Gradient borders](https://css-tricks.com/gradient-borders-in-css/)
- [ ] [JWK](https://tools.ietf.org/html/rfc7517)
- [ ] Support alternative auth providers e.g. auth0
- [ ] gitdb (?)
- [ ] [CRDT's](https://josephg.com/blog/crdts-are-the-future)

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