/* global */

import { update, deleted } from './tabular.js'
import { DB } from './db.js'

export function refreshed () {
  const groups = [...DB.groups.values()].sort((p, q) => p.index - q.index)

  realize(groups)

  groups.forEach(o => {
    const row = updateFromDB(o.OID, o)
    if (row) {
      if (o.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })

  DB.refreshed('groups')
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#groups tr[data-oid='" + oid + "']")

  if (record.status === 'deleted' || !row) {
    return
  }

  const name = row.querySelector(`[data-oid="${oid}.1"]`)

  row.dataset.status = record.status

  update(name, record.name)

  return row
}

function add (oid, record) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('groups').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#group')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('group')
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
      { suffix: 'name', oid: `${oid}.1`, selector: 'td input.name', flag: 'td img.name' }
    ]

    // const doors = [...DB.doors.values()].filter(o => o.status && o.status !== '<new>' && o.status !== 'deleted')
    //
    // doors.forEach(o => {
    // const m = o.OID.match(/^0\.4\.([1-9][0-9]*)$/)
    // const did = m[1]

    // record.doors.forEach((v, k) => {
    // if (v.group === g.OID) {
    //     fields.push({
    //       suffix: `g${gid}`,
    //       oid: `${k}`,
    //       selector: `td input.g${gid}`,
    //       flag: `td img.g${gid}`
    //     })
    //   }
    // })
    // })

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

function realize (groups) {
  groups.forEach(o => {
    const row = document.querySelector("div#groups tr[data-oid='" + o.oid + "']")

    if (o.status === 'deleted') {
      deleted('groups', row)
    } else if (!row) {
      add(o.OID, o)
    }
  })

  // ... doors
  const columns = document.querySelectorAll('.colheader.doorh')
  const cols = new Map([...columns].map(c => [c.dataset.door, c]))

  const doors = [...DB.doors.values()]
    .filter(o => o.status && o.status !== '<new>' && o.status !== 'deleted')
    .sort((p, q) => p.created.localeCompare(q.created))

  const missing = doors.filter(o => {
    return o.OID === '' || !cols.has(o.OID)
  })

  // // FIXME O(N²)
  // const surplus = [...columns].filter(c => {
  //   for (const d of doors) {
  //     if (cols.has(d.OID)) {
  //       return false
  //     }
  //   }
  //
  //   return true
  // })
  //
  // console.log('SURPLUS', surplus)

  const table = document.querySelector('#groups table')
  const thead = table.tHead
  const tbody = table.tBodies[0]

  missing.forEach(o => {
    const door = o.OID.match(/^0\.2\.([1-9][0-9]*)$/)[1]
    const template = document.querySelector('#door')
    const th = thead.rows[0].lastElementChild
    const padding = thead.rows[0].appendChild(document.createElement('th'))

    padding.classList.add('colheader')
    padding.classList.add('padding')

    th.classList.replace('padding', 'doorh')
    th.dataset.door = o.OID
    th.innerHTML = o.name

    for (const row of tbody.rows) {
      const uuid = row.id
      const oid = row.dataset.oid + '.X.' + door
      const ix = row.cells.length - 1
      const cell = row.insertCell(ix)

      cell.dataset.door = o.OID
      cell.innerHTML = template.innerHTML

      const flag = cell.querySelector('.flag')
      const field = cell.querySelector('.field')

      flag.classList.add(`d${door}`)
      field.classList.add(`d${door}`)

      flag.id = 'F' + oid

      field.id = uuid + '-' + `d${door}`
      field.dataset.oid = oid
      field.dataset.record = uuid
      field.dataset.original = ''
      field.dataset.value = ''
      field.checked = false
    }
  })
}
