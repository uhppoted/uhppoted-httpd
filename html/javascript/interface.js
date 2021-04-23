/* global */

import { postAsJSON, warning } from './uhppoted.js'
import { refreshed, busy, unbusy } from './system.js'
import { DB } from './db.js'

export function updateInterfaceFromDB (oid, record) {
  const section = document.querySelector(`[data-oid="${oid}"]`)

  if (section) {
    const name = section.querySelector(`[data-oid="${oid}.0"]`)
    const bind = section.querySelector(`[data-oid="${oid}.1"]`)
    const broadcast = section.querySelector(`[data-oid="${oid}.2"]`)
    const listen = section.querySelector(`[data-oid="${oid}.3"]`)

    update(name, record.name)
    update(bind, record.bind)
    update(broadcast, record.broadcast)
    update(listen, record.listen)
  }
}

export function rollbackx (tag, element) {
  const section = document.getElementById(tag)
  const oid = section.dataset.oid

  for (let id = 0; ; id++) {
    const element = section.querySelector(`[data-oid="${oid}.${id}"]`)
    if (element) {
      const flag = document.getElementById(`F${element.dataset.oid}`)

      element.dataset.value = element.dataset.original
      element.value = element.dataset.original
      element.classList.remove('modified')

      if (flag) {
        flag.classList.remove('modified')
        flag.classList.remove('pending')
      }

      continue
    }

    break
  }

  section.classList.remove('modified')
}

export function commitx (tag, element) {
  const section = document.getElementById(tag)
  const oid = section.dataset.oid
  const list = []

  for (let id = 0; ; id++) {
    const element = section.querySelector(`[data-oid="${oid}.${id}"]`)
    if (element) {
      if (element.dataset.value !== element.dataset.original) {
        list.push(element)
      }
      continue
    }
    break
  }

  list.forEach(e => {
    const flag = document.getElementById(`F${e.dataset.oid}`)
    if (flag) {
      flag.classList.remove('modified')
      flag.classList.add('pending')
    }
  })

  const records = []
  list.forEach(e => {
    const oid = e.dataset.oid
    const value = e.dataset.value
    records.push({ oid: oid, value: value })
  })

  const reset = function () {
    list.forEach(e => {
      const flag = document.getElementById(`F${e.dataset.oid}`)
      if (flag) {
        flag.classList.remove('pending')
        flag.classList.add('modified')
      }
    })
  }

  postx('objects', records, reset)

  list.forEach(e => {
    const flag = document.getElementById(`F${e.dataset.oid}`)
    if (flag) {
      flag.classList.remove('pending')
    }
  })
}

export function modifiedx (oid) {
  document.querySelector(`[data-oid="${oid}"]`)
  const container = document.querySelector(`[data-oid="${oid}"]`)
  let changed = false

  if (container) {
    for (let id = 0; ; id++) {
      const element = document.querySelector(`[data-oid="${oid}.${id}"]`)
      if (element) {
        changed = changed || element.classList.contains('modified')
        continue
      }

      break
    }

    if (changed) {
      container.classList.add('modified')
    } else {
      container.classList.remove('modified')
    }
  }
}

function postx (tag, records, reset) {
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
      unbusy()
    })
}

function update (element, value, status) {
  const v = value.toString()
  const oid = element.dataset.oid
  const flag = document.getElementById(`F${oid}`)

  if (element) {
    const previous = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently modified fields
    if (element.classList.contains('modified')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, flag)
      } else if (element.dataset.value !== v) {
        mark('modified', element, flag)
      } else {
        unmark('modified', element, flag)
        unmark('conflict', element, flag)
      }

      percolate(oid)
      return
    }

    //   // mark fields with unexpected values after submit

    //   if (td && td.classList.contains('pending')) {
    //     if (element.dataset.value !== v.toString()) {
    //       td.classList.add('conflict')
    //     } else {
    //       td.classList.remove('conflict')
    //     }
    //   }

    // update unmodified fields

    element.value = v

    setx('interface', element, value)
  }
}

export function setx (tag, element, value, status) {
  const oid = element.dataset.oid
  const original = element.dataset.original
  const v = value.toString()
  const flag = document.getElementById(`F${oid}`)

  element.dataset.value = v

  if (v !== original) {
    mark('modified', element, flag)
  } else {
    unmark('modified', element, flag)
  }

  percolate(oid)
}

function mark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.add(clazz)
    }
  })
}

function unmark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.remove(clazz)
    }
  })
}

function percolate (oid) {
  let oidx = oid
  while (oidx) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(oidx)
    oidx = match ? match[1] : null
    if (oidx) {
      modifiedx(oidx)
    }
  }
}
