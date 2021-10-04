import { busy, unbusy, dismiss, warning, getAsJSON, postAsJSON } from './uhppoted.js'
import * as common from './tabular.js'
import * as controllers from './controllers.js'
import * as LAN from './interfaces.js'
import { DB } from './db.js'

export function refreshed () {
  LAN.refreshed()
  controllers.refreshed()
}

export function onEdited (tag, event) {
  switch (tag) {
    case 'interface':
      LAN.set(event.target, event.target.value)
      break

    case 'controller': {
      common.set(event.target, event.target.value)
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
        common.set(event.target, event.target.value)
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
      common.set(event.target, event.target.checked)
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

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
export function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}
