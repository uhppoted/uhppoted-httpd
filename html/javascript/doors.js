/* global */

import { busy, unbusy, dismiss, warning, getAsJSON, postAsJSON } from './uhppoted.js'
import { DB } from './db.js'

export function create () {
  const records = [{ oid: '<new>', value: '' }]
  const reset = function () {}
  const cleanup = function () {}

  post('objects', records, reset, cleanup)
}

export function commit (...rows) {
  const list = []

  rows.forEach(row => {
    const oid = row.dataset.oid
    const children = row.querySelectorAll(`[data-oid^="${oid}."]`)
    children.forEach(e => {
      if (e.dataset.value !== e.dataset.original) {
        list.push(e)
      }
    })
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
      unmark('pending', e, flag)
      mark('modified', e, flag)
    })
  }

  const cleanup = function () {
    list.forEach(e => {
      const flag = document.getElementById(`F${e.dataset.oid}`)
      unmark('pending', e, flag)
    })
  }

  list.forEach(e => {
    const flag = document.getElementById(`F${e.dataset.oid}`)
    mark('pending', e, flag)
    unmark('modified', e, flag)
  })

  post('objects', records, reset, cleanup)
}

export function rollback (row) {
  if (row && row.classList.contains('new')) {
    DB.delete('doors', row.dataset.oid)
    refreshed()
  } else {
    revert(row)
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
  getAsJSON('/doors')
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
      console.error(err)
    })
}

export function post (tag, records, reset, cleanup) {
  busy()

  postAsJSON('/doors', { [tag]: records })
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
  const list = []

  DB.doors.forEach(c => {
    list.push(c)
  })

  list.sort((p, q) => {
    if (p.created < q.created) {
      return -1
    }

    if (p.created < q.created) {
      return +1
    }

    return 0
  })

  list.forEach(d => {
    const row = updateFromDB(d.OID, d)
    if (row) {
      if (d.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })

  DB.refreshed('doors')
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

function updateFromDB (oid, record) {
  let row = document.querySelector("div#doors tr[data-oid='" + oid + "']")

  if (record.status === 'deleted') {
    deleted(row)
    return
  }

  if (!row) {
    row = add(oid)
  }

  const name = row.querySelector(`[data-oid="${oid}.1"]`)
  const controller = row.querySelector(`[data-oid="${oid}.0.2.2"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}.0.2.3"]`)
  const door = row.querySelector(`[data-oid="${oid}.0.2.4"]`)
  const delay = row.querySelector(`[data-oid="${oid}.2"]`)
  const mode = row.querySelector(`[data-oid="${oid}.3"]`)

  row.dataset.status = record.status

  update(name, record.name)
  update(controller, record.controller)
  update(deviceID, record.deviceID)
  update(door, record.door)
  update(delay, record.delay.configured, record.delay.status)
  update(mode, record.mode.configured, record.mode.status)

  return row
}

function add (oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('doors').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#door')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('door')
    row.classList.add('new')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const commit = row.querySelector('td span.commit')
    commit.id = uuid + '_commit'
    commit.dataset.record = uuid
    commit.dataset.enabled = 'false'

    const rollback = row.querySelector('td span.rollback')
    rollback.id = uuid + '_rollback'
    rollback.dataset.record = uuid
    rollback.dataset.enabled = 'false'

    const fields = [
      { suffix: 'name', oid: `${oid}.1`, selector: 'td input.name', flag: 'td img.name' },
      { suffix: 'controller', oid: `${oid}.0.2.2`, selector: 'td input.controller', flag: 'td img.controller' },
      { suffix: 'deviceID', oid: `${oid}.0.2.3`, selector: 'td input.deviceID', flag: 'td img.deviceID' },
      { suffix: 'doorID', oid: `${oid}.0.2.4`, selector: 'td input.doorID', flag: 'td img.doorID' },
      { suffix: 'delay', oid: `${oid}.2`, selector: 'td input.delay', flag: 'td img.delay' },
      { suffix: 'mode', oid: `${oid}.3`, selector: 'td select.mode', flag: 'td img.mode' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)
      const flag = row.querySelector(f.flag)

      if (field) {
        field.id = uuid + '-' + f.suffix
        field.value = ''
        field.dataset.oid = f.oid
        field.dataset.record = uuid
        field.dataset.original = ''
        field.dataset.value = ''

        flag.id = 'F' + f.oid
      } else {
        console.error(f)
      }
    })

    return row
  }
}

function update (element, value, status) {
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
      element.value = v
    }

    set(element, value, status)
  }
}

function revert (row) {
  const fields = row.querySelectorAll('.field')

  fields.forEach((item) => {
    item.value = item.dataset.original
    set(item, item.dataset.original)
  })

  row.classList.remove('modified')
}

function modified (oid) {
  const element = document.querySelector(`[data-oid="${oid}"]`)

  if (element) {
    const list = document.querySelectorAll(`[data-oid^="${oid}."]`)
    const re = /^\.[0-9]+$/
    let count = 0

    list.forEach(e => {
      if (e.classList.contains('modified')) {
        const oidx = e.dataset.oid
        if (oidx.startsWith(oid) && re.test(oidx.substring(oid.length))) {
          count = count + 1
        }
      }
    })

    if (count > 0) {
      element.dataset.modified = count > 1 ? 'multiple' : 'single'
      element.classList.add('modified')
    } else {
      element.dataset.modified = null
      element.classList.remove('modified')
    }
  }
}

function deleted (row) {
  const tbody = document.getElementById('doors').querySelector('table tbody')

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

function percolate (oid, f) {
  let oidx = oid

  while (oidx) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(oidx)
    oidx = match ? match[1] : null
    if (oidx) {
      f(oidx)
    }
  }
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
