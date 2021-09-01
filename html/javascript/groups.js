/* global */

// import { busy, unbusy, warning, getAsJSON, postAsJSON } from './uhppoted.js'
import { update, deleted } from './edit.js'
import { DB } from './db.js'

// export function create () {
//   const records = [{ oid: '<new>', value: '' }]
//   const reset = function () {}
//   const cleanup = function () {}

//   post('objects', records, reset, cleanup)
// }

// export function commit (...rows) {
//   const list = []

//   rows.forEach(row => {
//     const oid = row.dataset.oid
//     const children = row.querySelectorAll(`[data-oid^="${oid}."]`)
//     children.forEach(e => {
//       if (e.classList.contains('modified')) {
//         list.push(e)
//       }
//     })
//   })

//   const records = []
//   list.forEach(e => {
//     const oid = e.dataset.oid
//     const value = e.dataset.value
//     records.push({ oid: oid, value: value })
//   })

//   const reset = function () {
//     list.forEach(e => {
//       const flag = document.getElementById(`F${e.dataset.oid}`)
//       unmark('pending', e, flag)
//       mark('modified', e, flag)
//     })
//   }

//   const cleanup = function () {
//     list.forEach(e => {
//       const flag = document.getElementById(`F${e.dataset.oid}`)
//       unmark('pending', e, flag)
//     })
//   }

//   list.forEach(e => {
//     const flag = document.getElementById(`F${e.dataset.oid}`)
//     mark('pending', e, flag)
//     unmark('modified', e, flag)
//   })

//   post('objects', records, reset, cleanup)
// }

// export function post (tag, records, reset, cleanup) {
//   busy()

//   postAsJSON('/doors', { [tag]: records })
//     .then(response => {
//       if (response.redirected) {
//         window.location = response.url
//       } else {
//         switch (response.status) {
//           case 200:
//             response.json().then(object => {
//               if (object && object.system && object.system.objects) {
//                 DB.updated('objects', object.system.objects)
//               }

//               refreshed()
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
//       cleanup()
//       unbusy()
//     })
// }

export function refreshed () {
  const list = []

  DB.groups.forEach(c => {
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

  DB.refreshed('groups')
}

function updateFromDB (oid, record) {
  let row = document.querySelector("div#groups tr[data-oid='" + oid + "']")

  if (record.status === 'deleted') {
    deleted('groups', row)
    return
  }

  if (!row) {
    row = add(oid)
  }

  const name = row.querySelector(`[data-oid="${oid}.1"]`)

  row.dataset.status = record.status

  update(name, record.name)

  return row
}

function add (oid) {
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

