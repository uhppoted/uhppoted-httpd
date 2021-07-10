/* global */

import { getAsJSON, postAsJSON, dismiss, warning } from './uhppoted.js'
import * as controllers from './controllers.js'
import * as LAN from './interface.js'
import { DB } from './db.js'

export function onEdited (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.value)
      break

    case 'controllers': {
      controllers.setX(event.target, event.target.value)
      break
    }
  }
}

export function onEnter (event) {
  if (event.key === 'Enter') {
    controllers.setX(event.target, event.target.value)
  }
}

export function onTick (event) {
  controllers.setX(event.target, event.target.checked)
}

export function onCommit (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.commit('interface', event.target)
      break

    case 'controller':
      // controllers.commit(event.target.dataset.record)
      controllers.commitX('controller', event.target)
      break

    default:
      console.log(`onCommit('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onCommitAll (tag, event) {
  if (tag === 'controller') {
    const tbody = document.getElementById('controllers').querySelector('table tbody')
    if (tbody) {
      const rows = tbody.rows
      const list = []

      for (let i = 0; i < rows.length; i++) {
        const row = rows[i]
        if (row.classList.contains('modified') || row.classList.contains('new')) {
          list.push(row.id)
        }
      }

      // controllers.commit(...list)
    }
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
  if (tag === 'controller') {
    const tbody = document.getElementById('controllers').querySelector('table tbody')
    if (tbody) {
      const rows = tbody.rows
      for (let i = rows.length; i > 0; i--) {
        controllers.rollback(rows[i - 1])
      }
    }
  }
}

export function onNew (event) {
  const record = { id: 'U' + uuidv4() }
  const records = [record]
  const reset = function () {}

  post(records, reset)
}

export function onRefresh (event) {
  if (event && event.target && event.target.id === 'refresh') {
    busy()
    dismiss()
  }

  getAsJSON('/system')
    .then(response => {
      unbusy()

      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.controllers) {
                object.system.controllers.forEach(l => {
                  DB.updated('interface', l.interface)
                  DB.updated('controllers', Object.values(l.controllers))
                })
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

function post (records, reset) {
  busy()

  postAsJSON('/system', { controllers: records })
    .then(response => {
      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.added) {
                DB.added('controllers', Object.values(object.system.added))
              }

              if (object && object.system && object.system.updated) {
                DB.updated('controllers', Object.values(object.system.updated))
              }

              if (object && object.system && object.system.deleted) {
                DB.deleted('controllers', Object.values(object.system.deleted))
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
      unbusy()
    })
}

export function refreshed () {
  const interfaces = DB.interfaces
  const boards = DB.controllers

  interfaces.forEach(c => {
    LAN.updateFromDB(c.OID, c)
  })

  boards.forEach(c => {
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

  DB.refreshed('controllers')
}

export function busy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  windmill.dataset.count = (queued + 1).toString()
}

export function unbusy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  if (queued > 1) {
    windmill.dataset.count = (queued - 1).toString()
  } else {
    delete (windmill.dataset.count)
  }
}

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}
