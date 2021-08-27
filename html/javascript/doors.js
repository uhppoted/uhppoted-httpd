/* global */

import { busy, unbusy, warning, getAsJSON, postAsJSON } from './uhppoted.js'
import { update, revert, mark, unmark } from './edit.js'
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
      if (e.classList.contains('modified')) {
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

function refreshed () {
  const list = []

  DB.doors.forEach(c => {
    list.push(c)
  })

  list.sort((p, q) => {
    if (p.created < q.created) {
      return -1
    } else if (p.created > q.created) {
      return +1
    } else {
      return 0
    }
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
  update(delay, record.delay.delay, record.delay.status)
  update(mode, record.mode.mode, record.mode.status)

  // ... set placeholders for blank names
  name.placeholder = record.name !== '' ? '-' : `<D${oid}>`.replaceAll('.', '')

  // ... set tooltips for error'd values
  { const tooltip = row.querySelector(`[data-oid="${oid}.2"] + div.tooltip-content`)

    if (tooltip) {
      const p = tooltip.querySelector('p')
      const err = record.delay.err && record.delay.err !== '' ? record.delay.err : ''
      const enabled = !!(record.delay.err && record.delay.err !== '')

      p.innerHTML = err

      if (enabled) {
        tooltip.classList.add('enabled')
      } else {
        tooltip.classList.remove('enabled')
      }
    }
  }

  { const tooltip = row.querySelector(`[data-oid="${oid}.3"] + div.tooltip-content`)

    if (tooltip) {
      const p = tooltip.querySelector('p')
      const err = record.mode.err && record.mode.err !== '' ? record.mode.err : ''
      const enabled = !!(record.mode.err && record.mode.err !== '')

      p.innerHTML = err

      if (enabled) {
        tooltip.classList.add('enabled')
      } else {
        tooltip.classList.remove('enabled')
      }
    }
  }

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
