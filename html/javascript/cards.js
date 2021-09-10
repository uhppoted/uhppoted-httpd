import { unbusy, getAsJSON, warning } from './uhppoted.js'
import { update, deleted } from './tabular.js'
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

export function refreshed () {
  // ... groups
  const columns = document.querySelectorAll('.colheader.grouph')
  const groups = [...DB.groups.values()].filter(g => g.status && g.status !== '<new>' && g.status !== 'deleted')

  groups.sort((p, q) => {
    return p.index - q.index
  })

  // FIXME O(N²)
  const missing = groups.filter(g => {
    for (const v of columns) {
      if (v.dataset.group === g.OID) {
        return false
      }
    }

    return true
  })

  // FIXME O(N²)
  const surplus = [...columns].filter(c => {
    for (const g of groups) {
      if (c.dataset.group === g.OID) {
        return false
      }
    }

    return true
  })

  missing.forEach(g => {
    const gid = g.OID.match(/^0\.4\.([1-9][0-9]*)$/)[1]
    const table = document.querySelector('#cards table')
    const thead = table.tHead
    const tbody = table.tBodies[0]
    const template = document.querySelector('#group')
    const th = thead.rows[0].appendChild(document.createElement('th'))

    th.classList.add('colheader')
    th.classList.add('grouph')
    th.dataset.group = g.OID
    th.innerHTML = g.name

    for (const row of tbody.rows) {
      const uuid = row.id
      const oid = row.dataset.oid + '.5.' + gid
      const cell = row.insertCell(-1)

      cell.dataset.group = g.OID
      cell.innerHTML = template.innerHTML

      const flag = cell.querySelector('.flag')
      const field = cell.querySelector('.field')

      flag.classList.add(`g${gid}`)
      field.classList.add(`g${gid}`)

      flag.id = 'F' + oid

      field.id = uuid + '-' + `g${gid}`
      field.dataset.oid = oid
      field.dataset.record = uuid
      field.dataset.original = ''
      field.dataset.value = ''
      field.checked = false
    }
  })

  surplus.forEach(col => {
    const group = col.dataset.group
    const cells = document.querySelectorAll(`td[data-group="${group}"]`)

    col.remove()
    cells.forEach(cell => cell.remove())
  })

  // ... cards
  const list = []

  DB.cards.forEach(c => {
    list.push(c)
  })

  list.sort((p, q) => {
    return p.created.localeCompare(q.created)
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

  DB.refreshed('cards')
}

function updateFromDB (oid, record) {
  let row = document.querySelector("div#cards tr[data-oid='" + oid + "']")

  if (record.status === 'deleted') {
    deleted('cards', row)
    return
  }

  if (!row) {
    row = add(oid, record)
  }

  const name = row.querySelector(`[data-oid="${oid}.1"]`)
  const number = row.querySelector(`[data-oid="${oid}.2"]`)
  const from = row.querySelector(`[data-oid="${oid}.3"]`)
  const to = row.querySelector(`[data-oid="${oid}.4"]`)
  const groups = [...DB.groups.values()].filter(g => g.status && g.status !== '<new>' && g.status !== 'deleted')

  row.dataset.status = record.status

  update(name, record.name)
  update(number, record.number)
  update(from, record.from)
  update(to, record.to)

  groups.forEach(g => {
    const td = row.querySelector(`td[data-group="${g.OID}"]`)

    if (td) {
      const e = td.querySelector('.field')
      const g = record.groups.get(`${e.dataset.oid}`)

      update(e, g.member)
    }
  })

  return row
}

function add (oid, record) {
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

    const groups = [...DB.groups.values()].filter(g => g.status && g.status !== '<new>' && g.status !== 'deleted')

    groups.forEach(g => {
      const m = g.OID.match(/^0\.4\.([1-9][0-9]*)$/)
      const gid = m[1]

      record.groups.forEach((v, k) => {
        if (v.group === g.OID) {
          fields.push({
            suffix: `g${gid}`,
            oid: `${k}`,
            selector: `td input.g${gid}`,
            flag: `td img.g${gid}`
          })
        }
      })
    })

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
