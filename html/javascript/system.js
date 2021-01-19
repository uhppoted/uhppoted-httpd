/* global constants */

// import { getAsJSON, dismiss } from './uhppoted.js'
import { postAsJSON, warning } from './uhppoted.js'

export function onEdited (event) {
  set('controllers', event.target, event.target.value)
}

export function onTick (event) {
  set('controllers', event.target, event.target.checked)
}

export function onCommit (event, op) {
  if (op && op === 'add') {
    onAdd(event)
    return
  }

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
  throw Error('onCommitAll: NOT IMPLEMENTED')
  // const tbody = document.getElementById('cardholders').querySelector('table tbody')

  // if (tbody) {
  //   const rows = tbody.rows
  //   const list = []

  //   for (let i = 0; i < rows.length; i++) {
  //     const row = rows[i]

  //     if (row.classList.contains('modified') || row.classList.contains('new')) {
  //       list.push(row.id)
  //     }
  //   }

  //   onUpdate(...list)
  // }
}

export function onRollbackAll (event) {
  throw Error('onRollbackAll: NOT IMPLEMENTED')
  // const tbody = document.getElementById('cardholders').querySelector('table tbody')

  // if (tbody) {
  //   const rows = tbody.rows

  //   for (let i = 0; i < rows.length; i++) {
  //     const row = rows[i]

  //     if (row.classList.contains('new')) {
  //       onDelete(row.id)
  //     } else if (row.classList.contains('modified')) {
  //       onRevert(row.id)
  //     }
  //   }
  // }
}

export function onAdd (event) {
  throw Error('onAdd: NOT IMPLEMENTED')
  // const id = event.target.dataset.record
  // const row = document.getElementById(id)

  // if (row) {
  //   const [record, fields] = rowToRecord(id, row)

  //   const reset = function () {
  //     row.classList.add('new')
  //     row.classList.add('modified')
  //     fields.forEach(f => { apply(f, (c) => { c.classList.add('modified') }) })
  //   }

  //   busy()
  //   row.classList.remove('new')
  //   row.classList.remove('modified')
  //   fields.forEach(f => { apply(f, (c) => { c.classList.remove('modified') }) })
  //   fields.forEach(f => { apply(f, (c) => { c.classList.add('pending') }) })

  //   postAsJSON('/cardholders', { cardholders: [record] })
  //     .then(response => {
  //       if (response.redirected) {
  //         window.location = response.url
  //       } else {
  //         switch (response.status) {
  //           case 200:
  //             response.json().then(object => {
  //               updated(object.db.updated)
  //               deleted(object.db.deleted)
  //             })
  //             break

  //           default:
  //             reset()
  //             response.text().then(message => { warning(message) })
  //         }
  //       }
  //     })
  //     .catch(function (err) {
  //       reset()
  //       warning(`Error committing record (ERR:${err.message.toLowerCase()})`)
  //     })
  //     .finally(() => {
  //       unbusy()
  //       fields.forEach(f => { apply(f, (c) => { c.classList.remove('pending') }) })
  //     })
  // }
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
              // deleted(object.system.deleted)
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
      const id = uuid + '-' + i

      d.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<input id="' + id + '" class="field door" type="text" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" data-status="" placeholder="-" readonly />'
    }
  }
}

export function onRefresh (event) {
  throw Error('onRefresh: NOT IMPLEMENTED')
  // busy()
  // dismiss()

  // getAsJSON('/cardholders')
  //   .then(response => {
  //     unbusy()

  //     switch (response.status) {
  //       case 200:
  //         response.json().then(object => { refresh(object.db) })
  //         break

  //       default:
  //         response.text().then(message => { warning(message) })
  //     }
  //   })
  //   .catch(function (err) {
  //     console.log(err)
  //   })
}

// function refresh (db) {
//  throw 'refresh: NOT IMPLEMENTED'
//  // updated(Object.values(db.cardholders))
// }

function updated (list) {
  if (list) {
    list.forEach((record) => {
      const id = record.ID
      const row = document.getElementById(id)

      if (row) {
        row.classList.remove('new')
      }

      if (record.Name) {
        update(document.getElementById(id + '-name'), record.Name)
      }
    })
  }
}

// function deleted (list) {
//   throw 'deleted: NOT IMPLEMENTED'
//   // const tbody = document.getElementById('cardholders').querySelector('table tbody')
//
//   // if (tbody && list) {
//   //   list.forEach((record) => {
//   //     const id = record.ID
//   //     const row = document.getElementById(id)
//
//   //     if (row) {
//   //       const rows = tbody.rows
//   //       for (let i = 0; i < rows.length; i++) {
//   //         if (rows[i].id === id) {
//   //           tbody.deleteRow(i)
//   //           break
//   //         }
//   //       }
//   //     }
//   //   })
//   // }
// }

function set (div, element, value) {
  const tbody = document.getElementById(div).querySelector('table tbody')
  const rowid = element.dataset.record
  const row = document.getElementById(rowid)
  const original = element.dataset.original
  const v = value.toString()

  element.dataset.value = v

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

function update (element, value) {
  const v = value.toString()

  if (element) {
    const td = cell(element)
    element.dataset.original = v

    // check for conflicts with concurrently modified fields

    if (td && td.classList.contains('modified')) {
      if (element.dataset.value !== v.toString()) {
        td.classList.add('conflict')
      } else {
        td.classList.remove('modified')
        td.classList.remove('conflict')
      }

      return
    }

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

    set('controllers', element, value)
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
  const fields = []

  const record = {
    id: id,
    oid: oid
  }

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

  return [record, fields]
}

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}
