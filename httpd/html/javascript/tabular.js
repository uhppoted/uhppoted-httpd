import * as overview from './overview.js'
import * as LAN from './interfaces.js'
import * as controllers from './controllers.js'
import * as doors from './doors.js'
import * as cards from './cards.js'
import * as groups from './groups.js'
import * as events from './events.js'
import * as logs from './logs.js'
import * as users from './users.js'
import { DB } from './db.js'
import { Cache } from './cache.js'
import { busy, unbusy, warning, dismiss, getAsJSON, postAsJSON } from './uhppoted.js'

class Warning extends Error {
  constructor (...params) {
    super(...params)

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, Warning)
    }

    this.name = 'Warning'
  }
}

HTMLTableSectionElement.prototype.sort = function (cb) {
  Array
    .prototype
    .slice
    .call(this.rows)
    .sort(cb)
    .forEach((e) => { this.appendChild(this.removeChild(e)) }, this)
}

const pages = {
  overview: {
    get: ['/controllers', '/doors', '/events?range=' + encodeURIComponent('0,15'), '/logs?range=' + encodeURIComponent('0,15')],
    refreshed: overview.refreshed
  },

  interfaces: {
    get: ['/interfaces'],
    post: '/interfaces',
    refreshed: LAN.refreshed
  },

  controllers: {
    get: ['/interfaces', '/controllers', '/doors'],
    post: '/controllers',
    refreshed: function () {
      LAN.refreshed()
      controllers.refreshed()
    },
    deletable: controllers.deletable
  },

  doors: {
    get: ['/doors', '/controllers'],
    post: '/doors',
    refreshed: doors.refreshed,
    deletable: doors.deletable
  },

  cards: {
    get: ['/cards', '/groups'],
    post: '/cards',
    refreshed: cards.refreshed,
    deletable: cards.deletable
  },

  groups: {
    get: ['/groups', '/doors'],
    post: '/groups',
    refreshed: groups.refreshed,
    deletable: groups.deletable
  },

  events: {
    get: ['/events?range=' + encodeURIComponent('0,15')],
    post: '/events',
    recordset: DB.events(),
    refreshed: events.refreshed
  },

  logs: {
    get: ['/logs?range=' + encodeURIComponent('0,15')],
    post: '/logs',
    recordset: DB.logs(),
    refreshed: logs.refreshed
  },

  users: {
    get: ['/users'],
    post: '/users',
    refreshed: users.refreshed,
    deletable: users.deletable
  }
}

export function onEdited (tag, event) {
  const status = event.target.dataset.status

  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.value)
      break

    case 'controller':
      set(event.target, event.target.value)
      break

    case 'door':
      set(event.target, event.target.value, status)

      // Allow 'forced' controller update for an error'd door
      if (status === 'error') {
        const element = event.target
        const tr = row(event.target)
        const td = cell(element)
        const oid = element.dataset.oid

        if (tr) {
          mark('modified', element, td)
          percolate(oid)
        }
      }
      break

    case 'card':
      set(event.target, event.target.value)
      break

    case 'group':
      set(event.target, event.target.value)
      break

    case 'user':
      set(event.target, event.target.value)
      break
  }
}

export function onEnter (tag, event) {
  const element = event.target
  const tr = row(element)
  const td = cell(element)
  const oid = element.dataset.oid
  const status = event.target.dataset.status

  if (event.key === 'Enter') {
    switch (tag) {
      case 'interface':
        LAN.set(element, element.value)
        break

      case 'controller':
        set(element, element.value)
        break

      case 'door':
        set(element, element.value, status)

        // Allow 'forced' controller update for an errored door mode/delay
        if (status === 'error') {
          if (tr) {
            mark('modified', element, td)
            percolate(oid)
          }
        }
        break

      case 'card':
        set(element, element.value)

        // Allow 'forced' controller update for an errored card
        if (status === 'error' || (tr && tr.dataset.status === 'error')) {
          mark('modified', element, td)
          percolate(oid)
        }
        break

      case 'group':
        set(element, element.value)
        break

      case 'user':
        set(element, element.value)
        break
    }
  }
}

export function onMore (tag, event) {
  switch (tag) {
    case 'events':
      more(pages.events)
      break

    case 'logs':
      more(pages.logs)
      break
  }
}

export function onTick (tag, event) {
  switch (tag) {
    case 'controller':
      set(event.target, event.target.checked ? 'tcp' : 'udp')
      break

    case 'door':
      set(event.target, event.target.checked ? 'true' : 'false')
      break

    case 'card':
      set(event.target, event.target.checked ? 'true' : 'false')
      break

    case 'group':
      set(event.target, event.target.checked ? 'true' : 'false')
      break

    case 'user':
      set(event.target, event.target.checked ? 'true' : 'false')
      break
  }
}

export function onCommit (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)
  let page

  switch (tag) {
    case 'interface':
      commit(pages.interfaces, LAN.changeset(event.target))
      break

    case 'controller':
    case 'door':
    case 'card':
    case 'group':
    case 'user':
      page = getPage(tag)
      commit(page, changeset(page, row))
      break
  }
}

export function onCommitAll (tag, event, table) {
  const tbody = document.getElementById(table).querySelector('table tbody')
  const rows = tbody.rows
  const list = []
  let page

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i]
    if (row.classList.contains('modified') || row.classList.contains('new')) {
      list.push(row)
    }
  }

  switch (tag) {
    case 'controllers':
    case 'doors':
    case 'cards':
    case 'groups':
    case 'users':
      page = getPage(tag)
      commit(page, changeset(page, ...list))
      break
  }
}

export function onRollback (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'interface':
      LAN.rollback('interface', event.target)
      break

    case 'controller':
      rollback('controllers', row, controllers.refreshed)
      break

    case 'door':
      rollback('doors', row, doors.refreshed)
      break

    case 'card':
      rollback('cards', row, cards.refreshed)
      break

    case 'group':
      rollback('groups', row, groups.refreshed)
      break

    case 'user':
      rollback('users', row, users.refreshed)
      break
  }
}

export function onRollbackAll (tag, event) {
  const start = Date.now()
  const cache = new Cache({ modified: false })
  const options = {
    cache: cache
  }

  const f = function (table, recordset, refreshed) {
    const rows = document.getElementById(table).querySelectorAll('table tbody tr:is(.modified,.new)')

    for (let i = rows.length; i > 0; i--) {
      rollback(tag, rows[i - 1], refreshed, options)
    }
  }

  switch (tag) {
    case 'controllers':
      f('controllers', 'controllers', function () {
        LAN.refreshed()
        controllers.refreshed()
      })
      break

    case 'doors':
      f('doors', 'doors', doors.refreshed)
      break

    case 'cards':
      f('cards', 'cards', cards.refreshed)
      break

    case 'groups':
      f('groups', 'groups', groups.refreshed)
      break

    case 'users':
      f('users', 'users', users.refreshed)
      break
  }

  console.log(`cards:rolled-back (${Date.now() - start}ms)`)
}

export function onNew (tag, event) {
  switch (tag) {
    case 'controller':
      create(pages.controllers)
      break

    case 'door':
      create(pages.doors)
      break

    case 'card':
      create(pages.cards)
      break

    case 'group':
      create(pages.groups)
      break

    case 'user':
      create(pages.users)
      break
  }
}

export function onRefresh (tag, event) {
  const page = getPage(tag)

  if (page) {
    if (event && event.target && event.target.id === 'refresh') {
      busy()
      dismiss()
    }

    get(page.get, page.refreshed)
  }
}

export function mark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.add(clazz)
    }
  })
}

export function unmark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.remove(clazz)
    }
  })
}

export function set (element, value, status, options = {}) {
  const oid = element.dataset.oid
  const original = element.dataset.original
  const v = value.toString()
  const td = cell(element)

  element.dataset.value = v

  if (status) {
    element.dataset.status = status
  } else {
    element.dataset.status = ''
  }

  if (v !== original) {
    mark('modified', element, td)
  } else {
    unmark('modified', element, td)
  }

  percolate(oid, options)
}

export function revert (row, options = {}) {
  const fields = row.querySelectorAll('.field')

  fields.forEach((item) => {
    let [value, ok] = DB.get(item.dataset.oid)
    if (!ok) {
      value = item.dataset.original
    }

    if (item && item.type === 'checkbox') {
      item.checked = value === 'true'
    } else {
      item.value = value
    }

    set(item, value, item.dataset.status, options)
  })

  row.classList.remove('modified')
}

export function update (element, value, status, checked, options = {}) {
  if (element && value !== undefined) {
    const v = value.toString()
    const oid = element.dataset.oid
    const td = element.parentElement
    const previous = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently edited fields
    if (element.classList.contains('modified')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, td)
      } else if (element.dataset.value !== v) {
        unmark('conflict', element, td)
      } else {
        unmark('conflict', element, td)
        unmark('modified', element, td)
      }

      percolate(oid)
      return
    }

    // check for conflicts with concurrently submitted fields
    if (element.classList.contains('pending')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, td)
      } else {
        unmark('conflict', element, td)
      }

      return
    }

    // update fields not pending, modified or editing
    if (element !== document.activeElement) {
      switch (element.getAttribute('type').toLowerCase()) {
        case 'checkbox':
          if (checked != null) {
            element.checked = checked(v)
          } else {
            element.checked = (v === 'true')
          }
          break

        default:
          element.value = v
      }
    }

    set(element, value, status, options)
  }
}

/**
  * Updates the 'modified' flag for the page root OID without caching.
  *
  * Interim fix for the edge cases introduced by the caching introduced to optimize 'modified'
  * for not-so-small cards lists.
  *
  */
export function recount (root) {
  // const rows = Array.from(document.querySelectorAll(`tr[data-oid^="${root}"]`)).map((row) => row.dataset.oid)
  //
  // for (const row of rows) {
  //   modified(`${row}`)
  // }

  modified(`${root}`)
}

function query (oid, cache) {
  if (cache != null) {
    return cache.query(oid)
  }

  return document.querySelector(`[data-oid="${oid}"]`)
}

function queryModified (oid, cache) {
  if (cache != null) {
    return cache.queryModified(oid)
  }

  return document.querySelectorAll(`[data-oid^="${oid}."]:is(.modified,.new)`)
}

function modified (oid, options = {}) {
  const { cache = null, recount = true } = options
  const element = query(oid, cache)

  if (element) {
    // <tr> and 'new' ?
    if (element.nodeName === 'TR') {
      const page = pageForRow(element)

      if (element.classList.contains('new') && page && page.deletable(element)) {
        element.classList.add('newish')
      } else {
        element.classList.remove('newish')
      }
    }

    // ... update 'modified' hierarchy

    if (recount) {
      const list = queryModified(oid, cache)
      const set = new Set(Array.from(list)
        .map(e => e.dataset.oid)
        .filter(v => v.startsWith(oid)))

      // .. count the 'unique parent' OIDs
      const f = (p, q) => p.length > q.length
      const r = (acc, v) => {
        if (!acc.find(e => v.startsWith(e + '.'))) {
          acc.push(v)
        }

        return acc
      }

      const count = [...set].sort(f).reduce(r, []).length

      if (count > 1) {
        element.dataset.modified = 'multiple'
        element.classList.add('modified')
      } else if (count > 0) {
        element.dataset.modified = 'single'
        element.classList.add('modified')
      } else {
        element.dataset.modified = null
        element.classList.remove('modified')
      }
    }
  }
}

export function deleted (tag, row) {
  const tbody = document.getElementById(tag).querySelector('table tbody')

  if (tbody && row) {
    const rows = tbody.rows

    for (let ix = 0; ix < rows.length; ix++) {
      if (rows[ix].id === row.id) {
        tbody.deleteRow(ix)
        break
      }
    }
  }
}

export function trim (tag, objects, rows) {
  const list = new Set(objects.map(o => o.OID))
  const remove = []

  rows.forEach(row => {
    if (!list.has(row.dataset.oid)) {
      remove.push(row)
    }
  })

  remove.forEach(row => {
    deleted(tag, row)
  })
}

function percolate (oid, options) {
  let oidx = oid

  while (oidx) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(oidx)
    oidx = match ? match[1] : null
    if (oidx) {
      modified(oidx, options)
    }
  }
}

function get (urls, refreshed) {
  const promises = []

  urls.forEach(url => {
    promises.push(new Promise((resolve, reject) => {
      getAsJSON(url)
        .then(response => {
          unbusy()

          if (response.redirected) {
            window.location = response.url
          } else if (response.status === 200) {
            return response.json()
          } else {
            response.text().then(message => {
              reject(new Warning(message))
            })
          }
        })
        .then((resolved, rejected) => {
          if (resolved) {
            for (const k in resolved) {
              DB.updated(k, resolved[k])
            }

            resolve()
          }
        })
    }))
  })

  Promise.all(promises).then((resolved, rejected) => {
    if (resolved) {
      const timestamp = document.querySelector('footer #timestamp')

      if (timestamp) {
        timestamp.innerHTML = datetime(new Date())
      }

      refreshed()
    }
  }).catch(err => {
    if (err instanceof Warning) {
      warning(err.message)
    } else {
      console.error(err)
    }
  })
}

function rollback (recordset, row, refreshed, options) {
  if (row && row.classList.contains('new')) {
    DB.delete(recordset, row.dataset.oid)
    refreshed()
  } else {
    revert(row, options)
  }
}

function commit (page, recordset) {
  const elements = recordset.updated
  const created = []
  const updated = []
  const deleted = recordset.deleted

  elements.forEach(e => {
    const oid = e.dataset.oid
    const value = e.dataset.value
    updated.push({ oid: oid, value: value })
  })

  const reset = function () {
    elements.forEach(e => {
      const td = e.parentElement
      unmark('pending', e, td)
      mark('modified', e, td)
    })
  }

  const cleanup = function () {
    elements.forEach(e => {
      const td = e.parentElement
      unmark('pending', e, td)
    })
  }

  elements.forEach(e => {
    const td = e.parentElement
    mark('pending', e, td)
    unmark('modified', e, td)
  })

  post(page.post, created, updated, deleted, page.refreshed, reset, cleanup)
}

function changeset (page, ...rows) {
  const updated = []
  const deleted = []

  rows.forEach(row => {
    const oid = row.dataset.oid

    if (page.deletable && page.deletable(row)) {
      deleted.push(oid)
    } else {
      const children = row.querySelectorAll(`[data-oid^="${oid}."]`)
      children.forEach(e => {
        if (e.classList.contains('modified')) {
          updated.push(e)
        }
      })
    }
  })

  return {
    updated: updated,
    deleted: deleted
  }
}

function create (page) {
  const created = [{ oid: '<new>', value: '' }]
  const reset = function () {}
  const cleanup = function () {}

  post(page.post, created, null, null, page.refreshed, reset, cleanup)
}

function more (page) {
  if (page.recordset) {
    const N = page.recordset.size
    const url = page.post + '?range=' + encodeURIComponent(`${N},+15`)

    get([url], page.refreshed)
  }
}

function post (url, created, updated, deleted, refreshed, reset, cleanup) {
  busy()

  postAsJSON(url, { created: created, updated: updated, deleted: deleted })
    .then(response => {
      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              for (const k in object) {
                DB.updated(k, object[k])
              }

              refreshed()
            })
            break

          case 401:
            reset()
            response.text().then(message => { warning(message.toUpperCase()) })
            break

          default:
            reset()
            response.text().then(message => { warning(message) })
        }
      }
    })
    .catch(function (err) {
      reset()
      warning(`Error committing changes (ERR:${err.message.toLowerCase()})`)
    })
    .finally(() => {
      cleanup()
      unbusy()
    })
}

function getPage (tag) {
  switch (tag) {
    case 'overview':
      return pages.overview

    case 'controller':
    case 'controllers':
      return pages.controllers

    case 'door':
    case 'doors':
      return pages.doors

    case 'card':
    case 'cards':
      return pages.cards

    case 'group':
    case 'groups':
      return pages.groups

    case 'events':
      return pages.events

    case 'logs':
      return pages.logs

    case 'user':
    case 'users':
      return pages.users
  }

  return null
}

function pageForRow (row) {
  const list = [
    { tag: 'controller', page: pages.controllers },
    { tag: 'door', page: pages.doors },
    { tag: 'card', page: pages.cards },
    { tag: 'group', page: pages.groups },
    { tag: 'user', page: pages.users }
  ]

  for (const v of list) {
    if (row.classList.contains(v.tag)) {
      return v.page
    }
  }

  return null
}

/* Returns the enclosing <td> element for an input/checkbox field
 */
function cell (element) {
  const parent = element.parentElement

  if (parent && parent.nodeName === 'TD') {
    return parent
  } else if (parent && parent.nodeName === 'LABEL' && parent.parentElement && parent.parentElement.nodeName === 'TD') {
    return parent.parentElement
  }

  return null
}

/* Returns the enclosing <tr> element for an input/checkbox field
 */
function row (element) {
  const td = cell(element)
  const parent = td ? td.parentElement : null

  if (parent && parent.nodeName === 'TR') {
    return parent
  }

  return null
}

function datetime (time) {
  const df = new Intl.DateTimeFormat('default', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })

  const m = new Map(df.formatToParts(time).map(o => [o.type, o.value]))
  const year = m.get('year')
  const month = m.get('month')
  const day = m.get('day')
  const hour = m.get('hour')
  const minute = m.get('minute')
  const second = m.get('second')

  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}
