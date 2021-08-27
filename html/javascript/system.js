/* global */

import { busy, unbusy, dismiss, warning, getAsJSON, postAsJSON } from './uhppoted.js'
import * as controllers from './controllers.js'
import * as LAN from './interface.js'
import { DB } from './db.js'

export function onEdited (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.value)
      break

    case 'controller': {
      controllers.set(event.target, event.target.value)
      break
    }
  }
}

export function onEnter (tag, event) {
  if (event.key === 'Enter') {
    switch (tag) {
      case 'interface':
        LAN.set(event.target, event.target.value)
        break

      case 'controller': {
        controllers.set(event.target, event.target.value)
        break
      }
    }
  }
}

export function onTick (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.checked)
      break

    case 'controller': {
      controllers.set(event.target, event.target.checked)
      break
    }
  }
}

export function onCommit (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.commit(event.target)
      break

    case 'controller': {
      const id = event.target.dataset.record
      const row = document.getElementById(id)

      controllers.commit(row)
    }
      break

    default:
      console.log(`onCommit('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onCommitAll (tag, event) {
  switch (tag) {
    case 'controller': {
      const tbody = document.getElementById('controllers').querySelector('table tbody')
      if (tbody) {
        const rows = tbody.rows
        const list = []

        for (let i = 0; i < rows.length; i++) {
          const row = rows[i]
          if (row.classList.contains('modified') || row.classList.contains('new')) {
            list.push(row)
          }
        }

        controllers.commit(...list)
      }
    }
      break
  }
}

export function onRollback (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.rollback('interface', event.target)
      break

    case 'controller': {
      const id = event.target.dataset.record
      const row = document.getElementById(id)
      controllers.rollback(row)
      break
    }

    default:
      console.log(`onRollback('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onRollbackAll (tag, event) {
  switch (tag) {
    case 'controller': {
      const rows = document.getElementById('controllers').querySelector('table tbody').rows
      for (let i = rows.length; i > 0; i--) {
        controllers.rollback(rows[i - 1])
      }
      break
    }
  }
}

export function onNew (tag, event) {
  if (tag === 'controller') {
    controllers.onNew()
  }
}

export function onRefresh (event) {
  if (event && event.target && event.target.id === 'refresh') {
    busy()
    dismiss()
  }

  get()
}

export function get () {
  getAsJSON('/system')
    .then(response => {
      unbusy()

      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.objects) {
                DB.updated('objects', object.system.objects)
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
      console.log(err)
    })
}

export function post (tag, records, reset, cleanup) {
  busy()

  postAsJSON('/system', { [tag]: records })
    .then(response => {
      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.objects) {
                DB.updated('objects', object.system.objects)
              }

              refreshed()
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
      warning(`Error committing record (ERR:${err.message.toLowerCase()})`)
    })
    .finally(() => {
      cleanup()
      unbusy()
    })
}

export function refreshed () {
  // ... update interface section
  DB.interfaces.forEach(c => {
    LAN.updateFromDB(c.OID, c)
  })

  // ... update controllers
  const list = []

  DB.controllers.forEach(c => {
    list.push(c)
  })

  list.sort((p, q) => {
    return p.created.localeCompare(q.created)
  })

  list.forEach(c => {
    const row = controllers.updateFromDB(c.OID, c)
    if (row) {
      switch (c.status) {
        case 'new':
          row.classList.add('new')
          break

        default:
          row.classList.remove('new')
      }
    }
  })

  // ... mark and sweep the DB
  DB.refreshed('controllers')
}

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
export function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}
