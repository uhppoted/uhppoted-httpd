/* global */

import { getAsJSON, postAsJSON, dismiss, warning } from './uhppoted.js'
import { DB } from './db.js'

export function onEdited (tag, event) {
  switch (tag) {
    case 'interface':
      setx('interface', event.target, event.target.value)
      break

    case 'controllers':
      set('controllers', event.target, event.target.value)
      break
  }
}

export function onEnter (event) {
  if (event.key === 'Enter') {
    set('controllers', event.target, event.target.value)
  }
}

export function onTick (event) {
  set('controllers', event.target, event.target.checked)
}

export function onCommit (tag, event) {
  switch (tag) {
    case 'controller':
      commit(event.target.dataset.record)
      break;

    default:
      console.log(`onCommit('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onCommitAll (tag, event) {
  switch (tag) {
    case 'controller':
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
        commit(...list)        
      }
      break

    default:
      console.log(`onCommitAll('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onRollback (tag, event, op) {
  switch (tag) {
    case 'controller':
      const id = event.target.dataset.record
      const row = document.getElementById(id)
      rollback(row)
      break;

    default:
      console.log(`onRollback('${tag}', ...)::NOT IMPLEMENTED`)
  }
}

export function onRollbackAll (tag, event) {
  switch (tag) {
    case 'controller':
      const tbody = document.getElementById('controllers').querySelector('table tbody')
      if (tbody) {
        const rows = tbody.rows
        for (let i = rows.length; i > 0; i--) {
          rollback(rows[i - 1])
        }
      }    
      break

    default:
      console.log(`onRollbackAll('${tag}', ...)::NOT IMPLEMENTED`)
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

function commit (...list) {
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

function rollback (row) {
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

function refreshed () {
  const interfaces = DB.interfaces
  const controllers = DB.controllers

  interfaces.forEach(c => {
    updateInterfaceFromDB(c.OID, c)
  })

  controllers.forEach(c => {
    const row = updateControllerFromDB(c.OID, c)
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

function updateInterfaceFromDB (oid, record) {
  const section = document.querySelector(`section[data-oid="${oid}"]`)

  if (section) {
    const name = section.querySelector(`[data-oid="${oid}.0"]`)
    const bind = section.querySelector(`[data-oid="${oid}.1"]`)
    const broadcast = section.querySelector(`[data-oid="${oid}.2"]`)
    const listen = section.querySelector(`[data-oid="${oid}.3"]`)

    name.innerHTML = record.name
    name.dataset.original = record.bind

    bind.value = record.bind
    bind.dataset.original = record.bind
    
    broadcast.value = record.broadcast
    broadcast.dataset.original = record.broadcast

    listen.value = record.listen
    listen.dataset.original = record.listen
  }
}

function updateControllerFromDB (oid, record) {
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

function add (uuid) {
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

      set('controllers', item, item.dataset.original)
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

function setx (tag, element, value, status) {
  const section = document.getElementById(tag)
  const oid = element.dataset.oid
  const original = element.dataset.original
  const v = value.toString()

  element.dataset.value = v
  if (v !== original) {
    element.classList.add('modified')
  } else {
    element.classList.remove('modified')
  }

  let xoid = oid
  while (xoid) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(xoid)
    xoid = match ? match[1] : null
    if (xoid) {
      modified(xoid)
    }
  }
}

function modified(oid) {document.querySelector(`[data-oid="${oid}"]`)
  const container = document.querySelector(`[data-oid="${oid}"]`)
  let changed = false

  if (container) {
    for (let id=0;; id++) {
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

function set (div, element, value, status) {
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

    set('controllers', element, value, status)
  }
}

function busy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  windmill.dataset.count = (queued + 1).toString()
}

function unbusy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  if (queued > 1) {
    windmill.dataset.count = (queued - 1).toString()
  } else {
    delete (windmill.dataset.count)
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

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}

function rowID (oid) {
  return 'R' + oid.replaceAll(/[^0-9]/g, '')
}
