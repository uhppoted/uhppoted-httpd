// /* global */

import { postAsJSON, warning } from './uhppoted.js'
import { refreshed, busy, unbusy } from './system.js'
import { DB } from './db.js'

export function updateFromDB (oid, record) {
  let row = document.querySelector("div#controllers tr[data-oid='" + oid + "']")

  if (record.status === 'deleted') {
    deleted(row)
    return
  }

  if (!row) {
    row = add(rowID(oid))
  }

  const id = row.id
  const name = document.getElementById(id + '-name')
  const deviceID = document.getElementById(id + '-ID')
  const address = document.getElementById(id + '-IP')
  const datetime = document.getElementById(id + '-datetime')
  const cards = document.getElementById(id + '-cards')
  const events = document.getElementById(id + '-events')
  const door1 = document.getElementById(id + '-door-1')
  const door2 = document.getElementById(id + '-door-2')
  const door3 = document.getElementById(id + '-door-3')
  const door4 = document.getElementById(id + '-door-4')

  row.dataset.oid = oid
  row.dataset.status = record.status

  update(name, record.name)
  update(deviceID, record.deviceID)
  update(address, record.address.address, record.address.status)
  update(datetime, record.datetime.datetime, record.datetime.status)
  update(cards, record.cards.cards, record.cards.status)
  update(events, record.events.events)
  update(door1, record.doors[1])
  update(door2, record.doors[2])
  update(door3, record.doors[3])
  update(door4, record.doors[4])

  address.dataset.original = record.address.configured
  datetime.dataset.original = record.datetime.expected

  return row
}

export function set (element, value, status) {
  const div = 'controllers'
  const tbody = document.getElementById(div).querySelector('table tbody')
  const rowid = element.dataset.record
  const row = document.getElementById(rowid)
  const original = element.dataset.original
  const v = value.toString()

  element.dataset.value = v

  if (status !== undefined && element.dataset.original !== undefined) {
    element.dataset.status = status
  }

  if (v !== original) {
    apply(element, (c) => { c.classList.add('modified') })
  } else {
    apply(element, (c) => { c.classList.remove('modified') })
  }

  if (row) {
    const unmodified = Array.from(row.children).every(item => !item.classList.contains('modified'))
    if (unmodified) {
      row.classList.remove('modified')
    } else {
      row.classList.add('modified')
    }
  }

  if (tbody) {
    const rows = tbody.rows
    const commitall = document.getElementById('commitall')
    const rollbackall = document.getElementById('rollbackall')
    let count = 0

    for (let i = 0; i < rows.length; i++) {
      if (rows[i].classList.contains('modified') || rows[i].classList.contains('new')) {
        count++
      }
    }

    commitall.style.display = count > 1 ? 'block' : 'none'
    rollbackall.style.display = count > 1 ? 'block' : 'none'
  }
}

export function commit (...list) {
  const rows = []
  const records = []
  const fields = []

  list.forEach(id => {
    const row = document.getElementById(id)
    if (row) {
      const [record, f] = rowToRecord(id, row)

      rows.push(row)
      records.push(record)
      fields.push(...f)
    }
  })

  const reset = function () {
    rows.forEach(r => r.classList.add('modified'))
    fields.forEach(f => { apply(f, (c) => { c.classList.add('modified') }) })
  }

  rows.forEach(r => r.classList.remove('modified'))
  fields.forEach(f => { apply(f, (c) => { c.classList.remove('modified') }) })
  fields.forEach(f => { apply(f, (c) => { c.classList.add('pending') }) })

  post(records, reset)

  fields.forEach(f => {
    apply(f, (c) => { c.classList.remove('pending') })
  })
}

export function rollback (row) {
  if (row && row.classList.contains('new')) {
    DB.delete('controllers', row.dataset.oid)
    refreshed()
  } else {
    revert(row)
  }
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

export function add (uuid) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#controller')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('controller')
    row.classList.add('new')
    row.dataset.oid = ''
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
      { suffix: 'name', selector: 'td input.name' },
      { suffix: 'ID', selector: 'td input.ID' },
      { suffix: 'IP', selector: 'td input.IP' },
      { suffix: 'datetime', selector: 'td input.datetime' },
      { suffix: 'cards', selector: 'td input.cards' },
      { suffix: 'events', selector: 'td input.events' },
      { suffix: 'door-1', selector: 'td select.door1' },
      { suffix: 'door-2', selector: 'td select.door2' },
      { suffix: 'door-3', selector: 'td select.door3' },
      { suffix: 'door-4', selector: 'td select.door4' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)

      field.id = uuid + '-' + f.suffix
      field.value = ''
      field.dataset.record = uuid
      field.dataset.original = ''
      field.dataset.value = ''
    })

    return row
  }
}

function revert (row) {
  if (row) {
    const id = row.id
    const fields = row.querySelectorAll('.field')

    fields.forEach((item) => {
      if ((item.dataset.record === id) && (item.dataset.value !== item.dataset.original)) {
        switch (item.getAttribute('type').toLowerCase()) {
          case 'text':
          case 'number':
          case 'date':
            item.value = item.dataset.original
            break

          case 'checkbox':
            item.checked = item.dataset.original === 'true'
            break
        }
      }

      set(item, item.dataset.original)
    })

    row.classList.remove('modified')
  }
}

function deleted (row) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')

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

function update (element, value, status) {
  const v = value.toString()

  if (element) {
    const td = cell(element)
    const original = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently modified fields

    if (td && td.classList.contains('modified')) {
      if (original !== v.toString() && element.dataset.value !== v.toString()) {
        td.classList.add('conflict')
      } else if (element.dataset.value !== v.toString()) {
        td.classList.add('modified')
      } else {
        td.classList.remove('modified')
        td.classList.remove('conflict')
      }

      return
    }

    element.dataset.original = v

    // mark fields with unexpected values after submit

    if (td && td.classList.contains('pending')) {
      if (element.dataset.value !== v.toString()) {
        td.classList.add('conflict')
      } else {
        td.classList.remove('conflict')
      }
    }

    // update unmodified fields

    switch (element.getAttribute('type').toLowerCase()) {
      case 'text':
      case 'number':
      case 'date':
        element.value = v
        break

      case 'checkbox':
        element.checked = (v === 'true')
        break

      case 'select':
        break
    }

    set(element, value, status)
  }
}

function cell (element) {
  let td = element

  for (let i = 0; i < 10; i++) {
    if (td.tagName.toLowerCase() === 'td') {
      return td
    }

    td = td.parentElement
  }

  return null
}

function apply (element, f) {
  const td = cell(element)

  if (td) {
    f(td)
  }
}

function rowToRecord (id, row) {
  const oid = row.dataset.oid
  const name = row.querySelector('#' + id + '-name')
  const deviceID = row.querySelector('#' + id + '-ID')
  const ip = row.querySelector('#' + id + '-IP')
  const datetime = row.querySelector('#' + id + '-datetime')
  const doors = {
    1: row.querySelector('#' + id + '-door-1'),
    2: row.querySelector('#' + id + '-door-2'),
    3: row.querySelector('#' + id + '-door-3'),
    4: row.querySelector('#' + id + '-door-4')
  }

  const record = {
    id: id,
    oid: oid
  }

  const fields = []

  if (name && name.dataset.value !== name.dataset.original) {
    record.name = name.value
    fields.push(name)
  }

  if (deviceID) {
    const v = Number(deviceID.value)

    if (v > 0) {
      record.deviceID = v
      fields.push(deviceID)
    }
  }

  if (ip && ip.dataset.value !== ip.dataset.original) {
    record.ip = ip.value
    fields.push(ip)
  }

  if (datetime && datetime.dataset.value !== datetime.dataset.original) {
    record.datetime = datetime.value
    fields.push(datetime)
  }

  for (const [k, door] of Object.entries(doors)) {
    if (door && door.dataset.value !== door.dataset.original) {
      if (!record.doors) {
        record.doors = {}
      }
      record.doors[k] = door.value
      fields.push(door)
    }
  }

  return [record, fields]
}

function rowID (oid) {
  return 'R' + oid.replaceAll(/[^0-9]/g, '')
}
