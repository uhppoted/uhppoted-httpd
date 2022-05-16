import { trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'

const pagesize = 5

export function refreshed () {
  const entries = [...DB.logs().values()]
    .filter(l => alive(l))
    .sort((p, q) => q.timestamp.localeCompare(p.timestamp))

  realize(entries)

  // renders a 'page size' of log entries
  const f = function (offset) {
    let ix = offset
    let count = 0
    while (count < pagesize && ix < entries.length) {
      const o = entries[ix]
      const row = updateFromDB(o.OID, o)
      if (row) {
        if (o.status === 'new') {
          row.classList.add('new')
        } else {
          row.classList.remove('new')
        }
      }

      count++
      ix++
    }
  }

  // sorts the table rows by 'timestamp'
  const g = function () {
    const table = document.querySelector('#logs table')
    const tbody = table.tBodies[0]

    tbody.sort((p, q) => {
      const u = DB.logs().get(p.dataset.oid)
      const v = DB.logs().get(q.dataset.oid)

      return v.timestamp.localeCompare(u.timestamp)
    })
  }

  // hides/shows the 'more' button
  const h = function () {
    const table = document.querySelector('#logs table')
    const tfoot = table.tFoot
    const last = DB.lastLog()

    if (last && DB.logs().has(last)) {
      tfoot.classList.add('hidden')
    } else {
      tfoot.classList.remove('hidden')
    }
  }

  // initialises the rows asynchronously in small'ish chunks
  const chunk = offset => new Promise(resolve => {
    f(offset)
    resolve(true)
  })

  async function * render () {
    for (let ix = 0; ix < entries.length; ix += pagesize) {
      yield chunk(ix).then(() => ix)
    }
  }

  (async function loop () {
    for await (const _ of render()) {
      // empty
    }
  })()
    .then(() => g())
    .then(() => h())
    .catch(err => console.error(err))
}

function realize (logs) {
  const table = document.querySelector('#logs table')
  const tbody = table.tBodies[0]

  trim('logs', logs, tbody.querySelectorAll('tr.entry'))

  logs.forEach(o => {
    let row = tbody.querySelector(`tr[data-oid='${o.OID}']`)
    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('logs').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#entry')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('entry')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const commit = row.querySelector('td span.commit')
    if (commit) {
      commit.id = uuid + '_commit'
      commit.dataset.record = uuid
    }

    const rollback = row.querySelector('td span.rollback')
    if (rollback) {
      rollback.id = uuid + '_rollback'
      rollback.dataset.record = uuid
    }

    const fields = [
      { suffix: 'timestamp', oid: `${oid}${schema.logs.timestamp}`, selector: 'td input.timestamp' },
      { suffix: 'uid', oid: `${oid}${schema.logs.uid}`, selector: 'td input.uid' },
      { suffix: 'item', oid: `${oid}${schema.logs.item}`, selector: 'td input.item' },
      { suffix: 'item-id', oid: `${oid}${schema.logs.itemID}`, selector: 'td input.item-id' },
      { suffix: 'item-name', oid: `${oid}${schema.logs.itemName}`, selector: 'td input.item-name' },
      { suffix: 'item-field', oid: `${oid}${schema.logs.field}`, selector: 'td input.item-field' },
      { suffix: 'details', oid: `${oid}${schema.logs.details}`, selector: 'td input.details' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)

      if (field) {
        field.id = uuid + '-' + f.suffix
        field.value = ''
        field.dataset.oid = f.oid
        field.dataset.record = uuid
        field.dataset.original = ''
        field.dataset.value = ''
      } else {
        console.error(f)
      }
    })

    return row
  }
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#logs tr[data-oid='" + oid + "']")

  const timestamp = row.querySelector(`[data-oid="${oid}${schema.logs.timestamp}"]`)
  const uid = row.querySelector(`[data-oid="${oid}${schema.logs.uid}"]`)
  const item = row.querySelector(`[data-oid="${oid}${schema.logs.item}"]`)
  const itemID = row.querySelector(`[data-oid="${oid}${schema.logs.itemID}"]`)
  const itemName = row.querySelector(`[data-oid="${oid}${schema.logs.itemName}"]`)
  const itemField = row.querySelector(`[data-oid="${oid}${schema.logs.field}"]`)
  const details = row.querySelector(`[data-oid="${oid}${schema.logs.details}"]`)

  row.dataset.status = record.status

  update(timestamp, format(record.timestamp))
  update(uid, record.uid)
  update(item, record.item.type)
  update(itemID, record.item.ID)
  update(itemName, record.item.name.toLowerCase())
  update(itemField, record.item.field.toLowerCase())
  update(details, record.item.details)

  return row
}

function update (element, value) {
  if (element && value !== undefined) {
    element.value = value.toString()
  }
}

function format (timestamp) {
  const dt = Date.parse(timestamp)
  const fmt = function (v) {
    return v < 10 ? '0' + v.toString() : v.toString()
  }

  if (!isNaN(dt)) {
    const date = new Date(dt)
    const year = date.getFullYear()
    const month = fmt(date.getMonth() + 1)
    const day = fmt(date.getDate())
    const hour = fmt(date.getHours())
    const minute = fmt(date.getMinutes())
    const second = fmt(date.getSeconds())

    return `${year}-${month}-${day} ${hour}:${minute}:${second}`
  }

  return ''
}
