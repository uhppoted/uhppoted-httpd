## v0.7.x

### IN PROGRESS

- [ ] ACL
      - rework door/button states as map[uint8]bool
      - groups => rules => permissions

- [ ] MemDB
      - convert tables to maps

- [ ] Card holders
      - JS timeout (if e.g. httpd isn't running any more)
      - card number
        - unit test for memdb.update
        - commit multiple rows (so you can e.g. switch card numbers)
      - wrap ACL update in goroutine
      - flag modified/conflicted fields with e.g. small red rectangle (not overbearing and overloaded border)
        -- https://css-tricks.com/a-complete-guide-to-data-attributes
        -- dataset.state = none, modified, pending, conflict (DAG ? go back to previous on rollback?)

      - rework 'groups' as checkboxes
      - use internal DB rather than JS dataset (?)
      - wrap templating in a decent error handler
      - from
      - to
      - name
      - add
      - delete
      - audit trail
      - virtual DOM
      - commit-all
      - rollback-all
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