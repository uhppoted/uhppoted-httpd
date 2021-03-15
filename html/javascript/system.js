/* global constants */

import { getAsJSON, postAsJSON, dismiss, warning } from './uhppoted.js'

export function onEdited (event) {
  set('controllers', event.target, event.target.value)
}

export function onEnter (event) {
  if (event.key === 'Enter') {
    set('controllers', event.target, event.target.value)
  }
}

export function onTick (event) {
  set('controllers', event.target, event.target.checked)
}

export function onCommit (event) {
  onUpdate(event.target.dataset.record)
}

export function onRollback (event, op) {
  if (op && op === 'delete') {
    onDelete(event.target.dataset.record)
    return
  }

  onRevert(event.target.dataset.record)
}

export function onCommitAll (event) {
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

    onUpdate(...list)
  }
}

export function onRollbackAll (event) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody) {
    const rows = tbody.rows

    for (let i = 0; i < rows.length; i++) {
      const row = rows[i]

      if (row.classList.contains('new')) {
        onDelete(row.id)
      } else if (row.classList.contains('modified')) {
        onRevert(row.id)
      }
    }
  }
}

export function onUpdate (...list) {
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

  busy()

  rows.forEach(r => r.classList.remove('modified'))
  fields.forEach(f => { apply(f, (c) => { c.classList.remove('modified') }) })
  fields.forEach(f => { apply(f, (c) => { c.classList.add('pending') }) })

  postAsJSON('/system', { controllers: records })
    .then(response => {
      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              if (object && object.system && object.system.updated) {
                updated(object.system.updated)
              }

              if (object && object.system && object.system.deleted) {
                deleted(object.system.deleted)
              }
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
      fields.forEach(f => { apply(f, (c) => { c.classList.remove('pending') }) })
    })
}

export function onDelete (id) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')
  const row = document.getElementById(id)

  if (tbody && row) {
    const rows = tbody.rows

    for (let ix = 0; ix < rows.length; ix++) {
      if (rows[ix].id === id) {
        tbody.deleteRow(ix)
        break
      }
    }
  }
}

export function onRevert (id) {
  const row = document.getElementById(id)

  if (row) {
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

export function onNew (event) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody) {
    const uuid = 'U' + uuidv4()
    const row = tbody.insertRow()
    const name = row.insertCell()
    const device = row.insertCell()
    const ip = row.insertCell()
    const datetime = row.insertCell()
    const cards = row.insertCell()
    const events = row.insertCell()
    const doors = {
      1: row.insertCell(),
      2: row.insertCell(),
      3: row.insertCell(),
      4: row.insertCell()
    }

    row.id = uuid
    row.classList.add('new')
    row.classList.add('controller')
    row.dataset.status = 'unknown'

    name.style = 'display:flex; flex-direction:row;'
    name.classList.add('rowheader')
    name.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<input id="' + uuid + '-name" class="field name" type="text" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" placeholder="-" />' +
                     '<span class="control commit" id="' + uuid + '_commit" onclick="onCommit(event)" data-record="' + uuid + '" data-enabled="false">&#9745;</span>' +
                     '<span class="control rollback" id="' + uuid + '_rollback" onclick="onRollback(event, \'delete\')" data-record="' + uuid + '" data-enabled="false">&#9746;</span>'

    device.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                       '<input id="' + uuid + '-ID" class="field ID" type="number" min="0" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" placeholder="-" />'

    ip.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                   '<input id="' + uuid + '-IP" class="field IP" type="text" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" data-status="" placeholder="-" />'

    datetime.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                         '<input id="' + uuid + '-datetime" class="field datetime" type="text" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" data-status="" placeholder="-" readonly />'

    cards.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                      '<input id="' + uuid + '-cards" class="field cards" type="number" min="0" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" data-status="" placeholder="-" readonly />'

    events.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                       '<input id="' + uuid + '-events" class="field events" type="number" min="0" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" data-status="" placeholder="-" readonly />'

    for (let i = 1; i <= 4; i++) {
      const d = doors[i]
      const id = uuid + '-door-' + i
      let options = '<option value="" selected>-</option>'

      constants.doors.forEach(v => {
        options += '<option value="' + v.id + '">' + v.door + '</option>'
      })

      d.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<select id="' + id + '"' +
                            ' class="field door"' +
                            ' type="text"' +
                            ' value=""' +
                            ' onchange="onEdited(event)"' +
                            ' data-record="' + uuid + '"' +
                            ' data-original=""' +
                            ' data-value=""' +
                            ' data-status=""' +
                            ' placeholder="-">' + options + '</select>'
    }
  }
}

export function onRefresh (event) {
  busy()
  dismiss()

  getAsJSON('/system')
    .then(response => {
      unbusy()

      switch (response.status) {
        case 200:
          response.json().then(object => { refresh(object.system) })
          break

        default:
          response.text().then(message => { warning(message) })
      }
    })
    .catch(function (err) {
      console.log(err)
    })
}

function refresh (sys) {
  updated(Object.values(sys.Controllers))
}

function updated (controllers) {
  if (controllers) {
    controllers.forEach((record) => {
      const id = record.ID
      const row = document.getElementById(id)

      if (row) {
        row.classList.remove('new')
        row.dataset.status = statusToString(record.Status)
      }

      if (record.Name) {
        update(document.getElementById(id + '-name'), record.Name)
      }

      if (record.DeviceID) {
        update(document.getElementById(id + '-ID'), record.DeviceID)
      }

      if (record.IP) {
        let ip = ''

        if (record.IP.IP !== null) {
          ip = record.IP.Address
        }

        update(document.getElementById(id + '-IP'), ip, statusToString(record.IP.Status))

        if (document.getElementById(id + '-IP')) {
          document.getElementById(id + '-IP').dataset.original = record.IP.Configured
        }
      }

      if (record.SystemTime) {
        update(document.getElementById(id + '-datetime'), record.SystemTime.DateTime, record.SystemTime.Status)
      }

      if (record.Cards) {
        update(document.getElementById(id + '-cards'), record.Cards.Records, statusToString(record.Cards.Status))
      }

      if (record.Events) {
        update(document.getElementById(id + '-events'), record.Events)
      }

      if (record.Doors) {
        update(document.getElementById(id + '-door-1'), record.Doors[1])
        update(document.getElementById(id + '-door-2'), record.Doors[2])
        update(document.getElementById(id + '-door-3'), record.Doors[3])
        update(document.getElementById(id + '-door-4'), record.Doors[4])
      }
    })
  }
}

function deleted (list) {
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody && list) {
    list.forEach((record) => {
      const id = record.ID
      const row = document.getElementById(id)

      if (row) {
        const rows = tbody.rows
        for (let i = 0; i < rows.length; i++) {
          if (rows[i].id === id) {
            tbody.deleteRow(i)
            break
          }
        }
      }
    })
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

function statusToString (status) {
  switch (status) {
    case 1:
      return 'ok'

    case 2:
      return 'uncertain'

    case 3:
      return 'error'

    case 4:
      return 'unconfigured'
  }

  return 'unknown'
}
