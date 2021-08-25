/* global constants */

import { getAsJSON, postAsJSON, warning, dismiss } from './uhppoted.js'
import { update } from './edit.js'
import { DB } from './db.js'

export function get () {
  getAsJSON('/cards')
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

function refreshed () {
  const list = []

  DB.cards.forEach(c => {
    list.push(c)
  })

  // list.sort((p, q) => {
  //   if (p.created < q.created) {
  //     return -1
  //   } else if (p.created > q.created) {
  //     return +1
  //   } else {
  //     return 0
  //   }
  // })

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

  DB.refreshed('cards')
}

function updateFromDB (oid, record) {
  let row = document.querySelector("div#cards tr[data-oid='" + oid + "']")

  // if (record.status === 'deleted') {
  //   deleted(row)
  //   return
  // }

  if (!row) {
    row = add(oid)
  }

  const name = row.querySelector(`[data-oid="${oid}.1"]`)
  const number = row.querySelector(`[data-oid="${oid}.2"]`)
  const from = row.querySelector(`[data-oid="${oid}.3"]`)
  const to = row.querySelector(`[data-oid="${oid}.4"]`)

  row.dataset.status = record.status

  update(name, record.name)
  update(number, record.number)
  update(from, record.from)
  update(to, record.to)

  return row
}

function add (oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('cards').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#card')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('card')
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
      { suffix: 'number', oid: `${oid}.2`, selector: 'td input.number', flag: 'td img.number' },
      { suffix: 'from', oid: `${oid}.3`, selector: 'td input.from', flag: 'td img.from' },
      { suffix: 'to', oid: `${oid}.4`, selector: 'td input.to', flag: 'td img.to' }
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

/** OLD STUFF **/
export function onEdited (event) {
  set(event.target, event.target.value)
}

export function onTick (event) {
  set(event.target, event.target.checked)
}

export function onCommit (event) {
  onUpdate(event.target.dataset.record)
}

export function onCommitAll (event) {
  const tbody = document.getElementById('cardholders').querySelector('table tbody')

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

export function onRollback (event, op) {
  if (op && op === 'delete') {
    onDelete(event.target.dataset.record)
    return
  }

  onRevert(event.target.dataset.record)
}

export function onRollbackAll (event) {
  const tbody = document.getElementById('cardholders').querySelector('table tbody')

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

  postAsJSON('/cardholders', { cardholders: records })
    .then(response => {
      if (response.redirected) {
        window.location = response.url
      } else {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              updated(object.db.updated)
              deleted(object.db.deleted)
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
  const tbody = document.getElementById('cardholders').querySelector('table tbody')
  const row = document.getElementById(id)

  if (row) {
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

      set(item, item.dataset.original)
    })

    row.classList.remove('modified')
  }
}

export function onNew (event) {
  const tbody = document.getElementById('cardholders').querySelector('table tbody')

  if (tbody) {
    const row = tbody.insertRow()
    const name = row.insertCell()
    const card = row.insertCell()
    const from = row.insertCell()
    const to = row.insertCell()
    const groups = []
    const uuid = 'U' + uuidv4()

    // 'constants' is a global object initialised by the Go template
    for (let i = 0; i < constants.groups.length; i++) {
      groups.push(row.insertCell())
    }

    row.id = uuid
    row.classList.add('new')

    name.style = 'display:flex; flex-direction:row;'
    name.classList.add('rowheader')

    name.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<input id="' + uuid + '-name" class="field name" type="text" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" placeholder="-" />' +
                     '<span class="control commit" id="' + uuid + '_commit" onclick="onCommit(event)" data-record="' + uuid + '" data-enabled="false">&#9745;</span>' +
                     '<span class="control rollback" id="' + uuid + '_rollback" onclick="onRollback(event, \'delete\')" data-record="' + uuid + '" data-enabled="false">&#9746;</span>'

    card.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<input id="' + uuid + '-card" class="field cardnumber" type="number" min="0" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" placeholder="6152346" />'

    from.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                     '<input id="' + uuid + '-from" class="field from" type="date" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" required />'

    to.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                   '<input id="' + uuid + '-to" class="field to" type="date" value="" onchange="onEdited(event)" data-record="' + uuid + '" data-original="" data-value="" required />'

    for (let i = 0; i < groups.length; i++) {
      const g = groups[i]
      const id = uuid + '-' + constants.groups[i]

      g.innerHTML = '<img class="flag" src="images/' + constants.theme + '/corner.svg" />' +
                    '<label class="group">' +
                    '<input id="' + id + '" class="field" type="checkbox" onclick="onTick(event)" data-record="' + uuid + '" data-original="false" data-value="false" />' +
                    '<img class="no"  src="images/times-solid.svg" />' +
                    '<img class="yes" src="images/check-solid.svg" />' +
                    '</label>'
    }
  }
}

export function onRefresh (event) {
  busy()
  dismiss()

  getAsJSON('/cardholders')
    .then(response => {
      unbusy()

      switch (response.status) {
        case 200:
          response.json().then(object => { refresh(object.db) })
          break

        default:
          response.text().then(message => { warning(message) })
      }
    })
    .catch(function (err) {
      console.log(err)
    })
}

function refresh (db) {
  updated(Object.values(db.cardholders))
}

function updated (list) {
  if (list) {
    list.forEach((record) => {
      const id = record.ID
      const row = document.getElementById(id)

      if (row) {
        row.classList.remove('new')
      }

      if (record.Name) {
        updateX(document.getElementById(id + '-name'), record.Name)
      }

      if (record.Card) {
        updateX(document.getElementById(id + '-card'), record.Card)
      }

      if (record.From) {
        updateX(document.getElementById(id + '-from'), record.From)
      }

      if (record.To) {
        updateX(document.getElementById(id + '-to'), record.To)
      }

      Object.entries(record.Groups).forEach(([k, v]) => {
        updateX(document.getElementById(id + '-' + k), v)
      })
    })
  }
}

function deleted (list) {
  const tbody = document.getElementById('cardholders').querySelector('table tbody')

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

function set (element, value) {
  const tbody = document.getElementById('cardholders').querySelector('table tbody')
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

function updateX (element, value) {
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

    set(element, value)
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
  const name = row.querySelector('#' + id + '-name')
  console.log('### ', id, row, name)
  const card = row.querySelector('#' + id + '-card')
  const from = row.querySelector('#' + id + '-from')
  const to = row.querySelector('#' + id + '-to')
  const fields = []

  const record = {
    id: id,
    groups: {}
  }

  if (name && name.dataset.value !== name.dataset.original) {
    const field = row.querySelector('#' + id + '-name')
    record.name = field.value
    fields.push(field)
  }

  if (card && card.dataset.value !== card.dataset.original) {
    const field = row.querySelector('#' + id + '-card')
    record.card = Number(field.value)
    fields.push(field)
  }

  if (from && from.dataset.value !== from.dataset.original) {
    const field = row.querySelector('#' + id + '-from')
    record.from = field.value
    fields.push(field)
  }

  if (to && to.dataset.value !== to.dataset.original) {
    const field = row.querySelector('#' + id + '-to')
    record.to = field.value
    fields.push(field)
  }

  constants.groups.forEach((gid) => {
    const field = row.querySelector('#' + id + '-' + gid)
    if (field && field.dataset.value !== field.dataset.original) {
      record.groups[gid] = field.checked
      fields.push(field)
    }
  })

  return [record, fields]
}

// Ref. https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
function uuidv4 () {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}
