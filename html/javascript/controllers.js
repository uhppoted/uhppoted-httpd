import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'
import * as combobox from './datetime.js'
// import { timezones } from './timezones.js'

const dropdowns = new Map()

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
  const address = row.querySelector(`[data-oid="${oid}${schema.controllers.endpoint.address}"]`)
  const datetime = row.querySelector(`[data-oid="${oid}${schema.controllers.datetime.current}"]`)
  const cards = row.querySelector(`[data-oid="${oid}${schema.controllers.cards.count}"]`)
  const events = row.querySelector(`[data-oid="${oid}${schema.controllers.events.last}"]`)
  const door1 = row.querySelector(`[data-oid="${oid}${schema.controllers.door1}"]`)
  const door2 = row.querySelector(`[data-oid="${oid}${schema.controllers.door2}"]`)
  const door3 = row.querySelector(`[data-oid="${oid}${schema.controllers.door3}"]`)
  const door4 = row.querySelector(`[data-oid="${oid}${schema.controllers.door4}"]`)

  // ... populate door dropdowns
  const doors = [...DB.doors.values()]
    .filter(o => o.status && o.status !== '<new>' && alive(o))
    .sort((p, q) => p.created.localeCompare(q.created));

  [door1, door2, door3, door4].forEach(select => {
    const options = select.options
    let ix = 1

    doors.forEach(d => {
      const value = d.OID
      const label = d.name !== '' ? d.name : `<D${d.OID}>`.replaceAll('.', '')

      if (ix < options.length) {
        if (options[ix].value !== value) {
          options.add(new Option(label, value, false, false), ix)
        } else if (options[ix].label !== label) {
          options[ix].label = label
        }
      } else {
        options.add(new Option(label, value, false, false))
      }

      ix++
    })

    while (options.length > (doors.length + 1)) {
      options.remove(options.length - 1)
    }
  })

  // ... set record values
  row.dataset.status = record.status

  update(name, record.name)
  update(deviceID, record.deviceID)
  update(address, record.address.address, record.address.status)
  update(datetime, record.datetime.datetime, record.datetime.status)
  update(cards, record.cards.cards, record.cards.status)
  update(events, record.events.last)
  update(door1, record.doors[1])
  update(door2, record.doors[2])
  update(door3, record.doors[3])
  update(door4, record.doors[4])

  address.dataset.original = record.address.configured
  datetime.dataset.original = record.datetime.expected

  // .. initialise date/time picker
  const cb = dropdowns.get(`${oid}${schema.controllers.datetime.current}`)

  if (cb) {
    combobox.set(cb, Date.parse(record.datetime.datetime))
  }

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

    const commit = row.querySelector('td span.commit')
    commit.id = uuid + '_commit'
    commit.dataset.record = uuid
    commit.dataset.enabled = 'false'

    const rollback = row.querySelector('td span.rollback')
    rollback.id = uuid + '_rollback'
    rollback.dataset.record = uuid
    rollback.dataset.enabled = 'false'

    const fields = [
      { suffix: 'name', oid: `${oid}${schema.controllers.name}`, selector: 'td input.name', flag: 'td img.name' },
      { suffix: 'ID', oid: `${oid}${schema.controllers.deviceID}`, selector: 'td input.ID', flag: 'td img.ID' },
      { suffix: 'IP', oid: `${oid}${schema.controllers.endpoint.address}`, selector: 'td input.IP', flag: 'td img.IP' },
      { suffix: 'datetime', oid: `${oid}${schema.controllers.datetime.current}`, selector: 'td input.datetime', flag: 'td img.datetime' },
      { suffix: 'cards', oid: `${oid}${schema.controllers.cards.count}`, selector: 'td input.cards', flag: 'td img.cards' },
      { suffix: 'events', oid: `${oid}${schema.controllers.events.last}`, selector: 'td input.events', flag: 'td img.events' },
      { suffix: 'door-1', oid: `${oid}${schema.controllers.door1}`, selector: 'td select.door1', flag: 'td img.door1' },
      { suffix: 'door-2', oid: `${oid}${schema.controllers.door2}`, selector: 'td select.door2', flag: 'td img.door2' },
      { suffix: 'door-3', oid: `${oid}${schema.controllers.door3}`, selector: 'td select.door3', flag: 'td img.door3' },
      { suffix: 'door-4', oid: `${oid}${schema.controllers.door4}`, selector: 'td select.door4', flag: 'td img.door4' }
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

    // .. initialise date/time picker
    const cb = combobox.initialise(row.querySelector('td.combobox'))

    dropdowns.set(`${oid}${schema.controllers.datetime.current}`, cb)

    return row
  }
}

