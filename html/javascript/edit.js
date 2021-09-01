import * as doors from './doors.js'
import * as cards from './cards.js'
import * as groups from './groups.js'
import * as db from './db.js'
import { busy, unbusy, warning, dismiss, getAsJSON } from './uhppoted.js'

export function onEdited (tag, event) {
  switch (tag) {
    case 'door':
      set(event.target, event.target.value)
      break

    case 'card':
      set(event.target, event.target.value)
      break

    case 'group':
      set(event.target, event.target.value)
      break
  }
}

export function onEnter (tag, event) {
  if (event.key === 'Enter') {
    switch (tag) {
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
    }
  }
}

export function onChoose (tag, event) {
  console.log('onChoose', event)

  event.target.selectedIndex = -1
  // switch (tag) {
  //   case 'door': {
  //     set(event.target, event.target.checked)
  //     break
  //   }
  // }
}

export function onTick (tag, event) {
  switch (tag) {
    case 'card':
      set(event.target, event.target.checked ? 'true' : '')
      break
  }
}

export function onCommit (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'door':
      doors.commit(row)
      break

    case 'card':
      cards.commit(row)
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
    case 'doors':
      doors.commit(...list)
      break

    case 'cards':
      cards.commit(...list)
      break
  }
}

export function onRollback (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'door':
      rollback('doors', row, doors.refreshed)
      break

    case 'card':
      rollback('cards', row, cards.refreshed)
      break

    case 'group':
      rollback('groups', row, groups.refreshed)
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
    case 'doors':
      f('doors', 'doors', doors.refreshed)
      break

    case 'cards':
      f('cards', 'cards', doors.refreshed)
      break

    case 'groups':
      f('groups', 'groups', groups.refreshed)
      break
  }
}

export function onNew (tag, event) {
  switch (tag) {
    case 'door':
      doors.create()
      break
  }
}

export function onRefresh (tag, event) {
  if (event && event.target && event.target.id === 'refresh') {
    busy()
    dismiss()
  }

  switch (tag) {
    case 'doors':
      doors.get()
      break

    case 'cards':
      cards.get()
      break

    case 'groups':
      get('/groups', groups.refreshed)
      break
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
    item.value = item.dataset.original
    set(item, item.dataset.original, item.dataset.status)
  })

  row.classList.remove('modified')
}

export function update (element, value, status) {
  if (element && value) {
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
    const f = (p, q) => p.length < q.length
    const r = (acc, v) => {
      if (!acc.find(e => v.startsWith(e))) {
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

function get (url, refreshed) {
  getAsJSON(url)
    .then(response => {
      unbusy()

      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.objects) {
                db.DB.updated('objects', object.system.objects)
              }

              refreshed()
            })
            break

          default:
            response.text().then(message => { warning(message) })
        }
      }
    })
    .catch(function (err) {
      console.error(err)
    })
}

function rollback (recordset, row, refreshed) {
  if (row && row.classList.contains('new')) {
    db.DB.delete(recordset, row.dataset.oid)
    refreshed()
  } else {
    revert(row)
  }
}
