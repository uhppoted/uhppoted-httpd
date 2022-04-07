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
    deleted: controllers.deleted
  },

  doors: {
    get: ['/doors', '/controllers'],
    post: '/doors',
    refreshed: doors.refreshed,
    deleted: doors.deleted
  },

  cards: {
    get: ['/cards', '/groups'],
    post: '/cards',
    refreshed: cards.refreshed,
    deleted: cards.deleted
  },

  groups: {
    get: ['/groups', '/doors'],
    post: '/groups',
    refreshed: groups.refreshed,
    deleted: groups.deleted
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
    deleted: users.deleted
  }
}

export function onEdited (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.value)
      break

    case 'controller':
      set(event.target, event.target.value)
      break

    case 'door':
      set(event.target, event.target.value)
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
  if (event.key === 'Enter') {
    switch (tag) {
      case 'interface':
        LAN.set(event.target, event.target.value)
        break

      case 'controller':
        set(event.target, event.target.value)
        break

      case 'door':
        set(event.target, event.target.value, event.target.dataset.status)

        { // Handles the case where 'Enter' is pressed on a field
          // to 'accept' the actual value which is different from
          // the 'configured' value.
          const element = event.target
          const configured = element.dataset.configured
          const v = element.dataset.value
          const oid = element.dataset.oid
          const flag = document.getElementById(`F${oid}`)

          if (configured && v !== configured) {
            mark('modified', element, flag)
            percolate(oid, modified)
          }
        }
        break

      case 'group':
        set(event.target, event.target.value)
        break

      case 'user':
        set(event.target, event.target.value)
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
    case 'card':
      set(event.target, event.target.checked ? 'true' : 'false')
      break

    case 'group':
      set(event.target, event.target.checked ? 'true' : 'false')
      break
  }
}

export function onCommit (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'interface':
      commit(pages.interfaces, LAN.changeset(event.target))
      break

    case 'controller':
      commit(pages.controllers, changeset(pages.controllers, row))
      break

    case 'door':
      commit(pages.doors, changeset(pages.doors, row))
      break

    case 'card':
      commit(pages.cards, changeset(pages.cards, row))
      break

    case 'group':
      commit(pages.groups, changeset(pages.groups, row))
      break

    case 'user':
      commit(pages.users, changeset(pages.users, row))
      break
  }
}

export function onCommitAll (tag, event, table) {
  const tbody = document.getElementById(table).querySelector('table tbody')
  const rows = tbody.rows
  const list = []

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i]
    if (row.classList.contains('modified') || row.classList.contains('new')) {
      list.push(row)
    }
  }

  switch (tag) {
    case 'controllers':
      commit(pages.controllers, changeset(pages.controllers, ...list))
      break

    case 'doors':
      commit(pages.doors, changeset(pages.doors, ...list))
      break

    case 'cards':
      commit(pages.cards, changeset(pages.cards, ...list))
      break

    case 'groups':
      commit(pages.groups, changeset(pages.groups, ...list))
      break

    case 'users':
      commit(pages.users, changeset(pages.users, ...list))
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
  const f = function (table, recordset, refreshed) {
    const rows = document.getElementById(table).querySelector('table tbody').rows
    for (let i = rows.length; i > 0; i--) {
      rollback(tag, rows[i - 1], refreshed)
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

export function set (element, value, status) {
  const oid = element.dataset.oid
  const original = element.dataset.original
  const v = value.toString()
  const flag = document.getElementById(`F${oid}`)

  element.dataset.value = v

  if (status) {
    element.dataset.status = status
  } else {
    element.dataset.status = ''
  }

  if (v !== original) {
    mark('modified', element, flag)
  } else {
    unmark('modified', element, flag)
  }

  percolate(oid, modified)
}

export function revert (row) {
  const fields = row.querySelectorAll('.field')

  fields.forEach((item) => {
    if (item && item.type === 'checkbox') {
      item.checked = item.dataset.original === 'true'
    } else {
      item.value = item.dataset.original
    }

    set(item, item.dataset.original, item.dataset.status)
  })

  row.classList.remove('modified')
}

export function update (element, value, status) {
  if (element && value !== undefined) {
    const v = value.toString()
    const oid = element.dataset.oid
    const flag = document.getElementById(`F${oid}`)
    const previous = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently edited fields
    if (element.classList.contains('modified')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, flag)
      } else if (element.dataset.value !== v) {
        unmark('conflict', element, flag)
      } else {
        unmark('conflict', element, flag)
        unmark('modified', element, flag)
      }

      percolate(oid, modified)
      return
    }

    // check for conflicts with concurrently submitted fields
    if (element.classList.contains('pending')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, flag)
      } else {
        unmark('conflict', element, flag)
      }

      return
    }

    // update fields not pending, modified or editing
    if (element !== document.activeElement) {
      switch (element.getAttribute('type').toLowerCase()) {
        case 'checkbox':
          element.checked = (v === 'true')
          break

        default:
          element.value = v
      }
    }

    set(element, value, status)
  }
}

export function modified (oid) {
  const element = document.querySelector(`[data-oid="${oid}"]`)

  if (element) {
    const list = document.querySelectorAll(`[data-oid^="${oid}."]`)
    const set = new Set()

    list.forEach(e => {
      if (e.classList.contains('modified')) {
        const oidx = e.dataset.oid
        if (oidx.startsWith(oid)) {
          set.add(oidx)
        }
      }
    })

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

export function percolate (oid, f) {
  let oidx = oid

  while (oidx) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(oidx)
    oidx = match ? match[1] : null
    if (oidx) {
      f(oidx)
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

function rollback (recordset, row, refreshed) {
  if (row && row.classList.contains('new')) {
    DB.delete(recordset, row.dataset.oid)
    refreshed()
  } else {
    revert(row)
  }
}

function commit (page, recordset) {
  const records = []
  const updated = recordset.updated
  const deleted = recordset.deleted

  updated.forEach(e => {
    const oid = e.dataset.oid
    const value = e.dataset.value
    records.push({ oid: oid, value: value })
  })

  const reset = function () {
    updated.forEach(e => {
      const flag = document.getElementById(`F${e.dataset.oid}`)
      unmark('pending', e, flag)
      mark('modified', e, flag)
    })
  }

  const cleanup = function () {
    updated.forEach(e => {
      const flag = document.getElementById(`F${e.dataset.oid}`)
      unmark('pending', e, flag)
    })
  }

  updated.forEach(e => {
    const flag = document.getElementById(`F${e.dataset.oid}`)
    mark('pending', e, flag)
    unmark('modified', e, flag)
  })

  post(page, records, deleted, reset, cleanup)
}

function changeset (page, ...rows) {
  const updated = []
  const deleted = []

  rows.forEach(row => {
    const oid = row.dataset.oid

    if (page.deleted && page.deleted(row)) {
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
  const records = [{ oid: '<new>', value: '' }]
  const reset = function () {}
  const cleanup = function () {}

  post(page, records, null, reset, cleanup)
}

function more (page) {
  if (page.recordset) {
    const N = page.recordset.size
    const url = page.post + '?range=' + encodeURIComponent(`${N},+15`)

    get([url], page.refreshed)
  }
}

function post (page, updated, deleted, reset, cleanup) {
  busy()

  postAsJSON(page.post, { objects: updated, deleted: deleted })
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

              page.refreshed()
            })
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

    case 'controllers':
      return pages.controllers

    case 'doors':
      return pages.doors

    case 'cards':
      return pages.cards

    case 'groups':
      return pages.groups

    case 'events':
      return pages.events

    case 'logs':
      return pages.logs

    case 'users':
      return pages.users
  }

  return null
}
