import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'

const pagesize = 5

export function refreshed () {
  const cards = [...DB.cards.values()]
    .filter(c => alive(c))
    .sort((p, q) => p.created.localeCompare(q.created))

  realize(cards)

  // renders a 'page size' chunk of cards
  const f = function (offset) {
    let ix = offset
    let count = 0
    while (count < pagesize && ix < cards.length) {
      const o = cards[ix]
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

  // sorts the table rows by 'created'
  const g = function () {
    const focused = document.activeElement

    if (!focused || focused.nodeName !== 'INPUT') {
      const table = document.querySelector('#cards table')
      const tbody = table.tBodies[0]

      tbody.sort((p, q) => {
        const u = DB.cards.get(p.dataset.oid)
        const v = DB.cards.get(q.dataset.oid)

        return u.created.localeCompare(v.created)
      })
    }
  }

  const chunk = offset => new Promise(resolve => {
    f(offset)
    resolve(true)
  })

  async function * render () {
    for (let ix = 0; ix < cards.length; ix += pagesize) {
      yield chunk(ix).then(() => ix)
    }
  }

  (async function loop () {
    for await (const _ of render()) {
      // empty
    }
  })()
    .then(() => g())
    .catch(err => console.error(err))
}

export function deletable (row) {
  const name = row.querySelector('td input.name')
  const card = row.querySelector('td input.number')
  const re = /^\s*$/

  if (name && name.dataset.oid !== '' && card && card.dataset.oid !== '') {
    const v = parseInt(card.dataset.value, 10)

    return re.test(name.dataset.value) && (Number.isNaN(v) || v === 0)
  }

  return false
}

function realize (cards) {
  const table = document.querySelector('#cards table')
  const thead = table.tHead
  const tbody = table.tBodies[0]

  const groups = new Map([...DB.groups.values()]
    .filter(g => g.status && g.status !== '<new>' && alive(g))
    .sort((p, q) => p.created.localeCompare(q.created))
    .map(o => [o.OID, o]))

  // ... columns
  const columns = table.querySelectorAll('th.group')
  const cols = new Map([...columns].map(c => [c.dataset.group, c]))
  const missing = [...groups.values()].filter(o => o.OID === '' || !cols.has(o.OID))
  const surplus = [...cols].filter(([k]) => !groups.has(k))

  missing.forEach(o => {
    const th = thead.rows[0].lastElementChild
    const padding = thead.rows[0].appendChild(document.createElement('th'))

    padding.classList.add('colheader')
    padding.classList.add('padding')

    th.classList.replace('padding', 'group')
    th.dataset.group = o.OID
    th.innerHTML = o.name
  })

  surplus.forEach(([, v]) => {
    v.remove()
  })

  // ... rows
  trim('cards', cards, tbody.querySelectorAll('tr.card'))

  cards.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (!row) {
      row = add(o.OID, o)
    }

    const columns = row.querySelectorAll('td.group')
    const cols = new Map([...columns].map(c => [c.dataset.group, c]))
    const missing = [...groups.values()].filter(o => o.OID === '' || !cols.has(o.OID))
    const surplus = [...cols].filter(([k]) => !groups.has(k))

    missing.forEach(o => {
      const group = o.OID.match(schema.groups.regex)[2]
      const template = document.querySelector('#group')

      const uuid = row.id
      const oid = row.dataset.oid + `${schema.cards.group}` + group
      const ix = row.cells.length - 1
      const cell = row.insertCell(ix)

      cell.classList.add('group')
      cell.dataset.group = o.OID
      cell.innerHTML = template.innerHTML

      const field = cell.querySelector('.field')

      field.id = uuid + '-' + `g${group}`
      field.dataset.oid = oid
      field.dataset.record = uuid
      field.dataset.original = ''
      field.dataset.value = ''
      field.checked = false
    })

    surplus.forEach(([, v]) => {
      v.remove()
    })
  })
}

function add (oid, record) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('cards').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#card')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('card')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const commit = row.querySelector('td span.commit')
    commit.id = uuid + '_commit'
    commit.dataset.record = uuid

    const rollback = row.querySelector('td span.rollback')
    rollback.id = uuid + '_rollback'
    rollback.dataset.record = uuid

    const fields = [
      { suffix: 'name', oid: `${oid}${schema.cards.name}`, selector: 'td input.name' },
      { suffix: 'number', oid: `${oid}${schema.cards.card}`, selector: 'td input.number' },
      { suffix: 'PIN', oid: `${oid}${schema.cards.PIN}`, selector: 'td input.PIN' },
      { suffix: 'from', oid: `${oid}${schema.cards.from}`, selector: 'td input.from' },
      { suffix: 'to', oid: `${oid}${schema.cards.to}`, selector: 'td input.to' }
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
  const row = document.querySelector("div#cards tr[data-oid='" + oid + "']")

  const name = row.querySelector(`[data-oid="${oid}${schema.cards.name}"]`)
  const number = row.querySelector(`[data-oid="${oid}${schema.cards.card}"]`)
  const PIN = row.querySelector(`[data-oid="${oid}${schema.cards.PIN}"]`)
  const from = row.querySelector(`[data-oid="${oid}${schema.cards.from}"]`)
  const to = row.querySelector(`[data-oid="${oid}${schema.cards.to}"]`)
  const groups = [...DB.groups.values()].filter(g => g.status && g.status !== '<new>' && alive(g))

  row.dataset.status = record.status

  update(name, record.name)
  update(number, parseInt(record.number, 10) === 0 ? '' : record.number)
  update(PIN, parseInt(record.PIN, 10) === 0 ? '' : record.PIN)
  update(from, record.from)
  update(to, record.to)

  groups.forEach(g => {
    const td = row.querySelector(`td[data-group="${g.OID}"]`)

    if (td) {
      const e = td.querySelector('.field')
      const g = record.groups.get(`${e.dataset.oid}`)

      update(e, g && g.member)
    }
  })

  return row
}
