import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'

export function refreshed () {
  const list = [...DB.controllers.values()]
    .filter(c => alive(c))
    .sort((p, q) => p.created.localeCompare(q.created))

  realize(list)

  list.forEach(o => {
    const row = updateFromDB(o.OID, o)
    if (row) {
      if (o.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#controllers tr[data-oid='" + oid + "']")

  const name = row.querySelector(`[data-oid="${oid}${schema.controllers.name}"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}${schema.controllers.deviceID}"]`)
  const datetime = row.querySelector(`[data-oid="${oid}${schema.controllers.datetime.current}"]`)
  const cards = row.querySelector(`[data-oid="${oid}${schema.controllers.cards.count}"]`)
  const events = row.querySelector(`[data-oid="${oid}${schema.controllers.events.last}"]`)
  const door1 = row.querySelector(`[data-oid="${oid}${schema.controllers.door1}"]`)
  const door2 = row.querySelector(`[data-oid="${oid}${schema.controllers.door2}"]`)
  const door3 = row.querySelector(`[data-oid="${oid}${schema.controllers.door3}"]`)
  const door4 = row.querySelector(`[data-oid="${oid}${schema.controllers.door4}"]`)

  // ... set record values
  const doors = new Map([...DB.doors.values()].map(o => [o.OID, o.name]))

  row.dataset.status = record.status

  update(name, record.name)
  update(deviceID, record.deviceID)
  update(datetime, record.datetime.datetime, record.datetime.status)
  update(cards, record.cards.cards, record.cards.status)
  update(events, record.events.last)
  update(door1, doors.get(record.doors[1]))
  update(door2, doors.get(record.doors[2]))
  update(door3, doors.get(record.doors[3]))
  update(door4, doors.get(record.doors[4]))

  datetime.dataset.original = record.datetime.expected

  return row
}

function realize (controllers) {
  const table = document.querySelector('#controllers table')
  const tbody = table.tBodies[0]

  trim('controllers', controllers, tbody.querySelectorAll('tr.controller'))

  controllers.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid, record) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#controller')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('controller')
    row.classList.add('new')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const fields = [
      { suffix: 'name', oid: `${oid}${schema.controllers.name}`, selector: 'td input.name', flag: 'td img.name' },
      { suffix: 'ID', oid: `${oid}${schema.controllers.deviceID}`, selector: 'td input.ID', flag: 'td img.ID' },
      { suffix: 'datetime', oid: `${oid}${schema.controllers.datetime.current}`, selector: 'td input.datetime', flag: 'td img.datetime' },
      { suffix: 'cards', oid: `${oid}${schema.controllers.cards.count}`, selector: 'td input.cards', flag: 'td img.cards' },
      { suffix: 'events', oid: `${oid}${schema.controllers.events.last}`, selector: 'td input.events', flag: 'td img.events' },
      { suffix: 'door-1', oid: `${oid}${schema.controllers.door1}`, selector: 'td input.door1', flag: 'td img.door1' },
      { suffix: 'door-2', oid: `${oid}${schema.controllers.door2}`, selector: 'td input.door2', flag: 'td img.door2' },
      { suffix: 'door-3', oid: `${oid}${schema.controllers.door3}`, selector: 'td input.door3', flag: 'td img.door3' },
      { suffix: 'door-4', oid: `${oid}${schema.controllers.door4}`, selector: 'td input.door4', flag: 'td img.door4' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)
      const flag = row.querySelector(f.flag)

      field.id = uuid + '-' + f.suffix
      field.value = ''
      field.dataset.oid = f.oid
      field.dataset.record = uuid
      field.dataset.original = ''
      field.dataset.value = ''

      flag.id = 'F' + f.oid
    })

    return row
  }
}
