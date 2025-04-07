import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'
import { Combobox } from './combobox.js'
import { timezones } from './timezones.js'
import { loaded } from './uhppoted.js'

export function refreshed() {
  const list = [...DB.controllers.values()].filter((c) => alive(c)).sort((p, q) => p.created.localeCompare(q.created))

  realize(list)

  list.forEach((o) => {
    const row = updateFromDB(o.OID, o)
    if (row) {
      if (o.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })

  loaded()
}

export function deletable(row) {
  const name = row.querySelector('td input.name')
  const id = row.querySelector('td input.ID')
  const re = /^\s*$/

  if (name && name.dataset.oid !== '' && re.test(name.dataset.value) && id && id.dataset.oid !== '' && re.test(id.dataset.value)) {
    return true
  }

  return false
}

function realize(controllers) {
  const table = document.querySelector('#controllers table')
  const tbody = table.tBodies[0]

  trim('controllers', controllers, tbody.querySelectorAll('tr.controller'))

  controllers.forEach((o) => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add(oid, _record) {
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

    const rollback = row.querySelector('td span.rollback')
    rollback.id = uuid + '_rollback'
    rollback.dataset.record = uuid

    const fields = [
      {
        suffix: 'name',
        oid: `${oid}${schema.controllers.name}`,
        selector: 'td input.name',
      },
      {
        suffix: 'ID',
        oid: `${oid}${schema.controllers.deviceID}`,
        selector: 'td input.ID',
      },
      {
        suffix: 'IP',
        oid: `${oid}${schema.controllers.endpoint.address}`,
        selector: 'td input.IP',
      },
      {
        suffix: 'protocol',
        oid: `${oid}${schema.controllers.endpoint.protocol}`,
        selector: 'td label.protocol input',
      },
      {
        suffix: 'datetime',
        oid: `${oid}${schema.controllers.datetime.current}`,
        selector: 'td input.datetime',
      },
      {
        suffix: 'interlock',
        oid: `${oid}${schema.controllers.interlock}`,
        selector: 'td select.interlock',
      },
      {
        suffix: 'antipassback',
        oid: `${oid}${schema.controllers.antipassback.antipassback}`,
        selector: 'td select.antipassback',
      },
      {
        suffix: 'cards',
        oid: `${oid}${schema.controllers.cards.count}`,
        selector: 'td input.cards',
      },
      {
        suffix: 'events',
        oid: `${oid}${schema.controllers.events.last}`,
        selector: 'td input.events',
      },
      {
        suffix: 'door-1',
        oid: `${oid}${schema.controllers.door1}`,
        selector: 'td select.door1',
      },
      {
        suffix: 'door-2',
        oid: `${oid}${schema.controllers.door2}`,
        selector: 'td select.door2',
      },
      {
        suffix: 'door-3',
        oid: `${oid}${schema.controllers.door3}`,
        selector: 'td select.door3',
      },
      {
        suffix: 'door-4',
        oid: `${oid}${schema.controllers.door4}`,
        selector: 'td select.door4',
      },
    ]

    fields.forEach((f) => {
      const field = row.querySelector(f.selector)
      if (field) {
        field.id = uuid + '-' + f.suffix
        field.value = ''
        field.dataset.oid = f.oid
        field.dataset.record = uuid
        field.dataset.original = ''
        field.dataset.value = ''

        // ... sigh .. Safari is awful
        if (`${navigator.vendor}`.toLowerCase().includes('apple')) {
          field.classList.add('apple')
        }
      }
    })

    // .. initialise date/time picker
    combobox(row.querySelector('td.combobox'))

    return row
  }
}

function updateFromDB(oid, record) {
  const row = document.querySelector("div#controllers tr[data-oid='" + oid + "']")

  const name = row.querySelector(`[data-oid="${oid}${schema.controllers.name}"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}${schema.controllers.deviceID}"]`)
  const address = row.querySelector(`[data-oid="${oid}${schema.controllers.endpoint.address}"]`)
  const protocol = row.querySelector(`[data-oid="${oid}${schema.controllers.endpoint.protocol}"]`)
  const datetime = row.querySelector(`[data-oid="${oid}${schema.controllers.datetime.current}"]`)
  const interlock = row.querySelector(`[data-oid="${oid}${schema.controllers.interlock}"]`)
  const antipassback = row.querySelector(`[data-oid="${oid}${schema.controllers.antipassback.antipassback}"]`)
  const cards = row.querySelector(`[data-oid="${oid}${schema.controllers.cards.count}"]`)
  const events = row.querySelector(`[data-oid="${oid}${schema.controllers.events.last}"]`)
  const door1 = row.querySelector(`[data-oid="${oid}${schema.controllers.door1}"]`)
  const door2 = row.querySelector(`[data-oid="${oid}${schema.controllers.door2}"]`)
  const door3 = row.querySelector(`[data-oid="${oid}${schema.controllers.door3}"]`)
  const door4 = row.querySelector(`[data-oid="${oid}${schema.controllers.door4}"]`)

  // ... populate door dropdowns
  const doors = [...DB.doors.values()]
    .filter((o) => o.status && o.status !== '<new>' && alive(o))
    .sort((p, q) => p.created.localeCompare(q.created))

  ;[door1, door2, door3, door4].forEach((select) => {
    const options = select.options

    let ix = 1

    doors.forEach((d) => {
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

    while (options.length > doors.length + 1) {
      options.remove(options.length - 1)
    }
  })

  // ... set record values
  row.dataset.status = record.status

  const dt = record.datetime.status === 'uncertain' ? record.datetime.configured : record.datetime.datetime

  update(name, record.name)
  update(deviceID, record.deviceID)
  update(address, record.address.address, record.address.status)
  update(protocol, record.protocol === 'tcp' ? 'tcp' : 'udp', null, (v) => {
    return v === 'tcp'
  })
  update(datetime, dt, record.datetime.status)
  update(interlock, record.interlock)
  update(antipassback, record.antipassback)
  update(cards, record.cards.cards, record.cards.status)
  update(events, record.events.last, record.events.status)
  update(door1, record.doors[1])
  update(door2, record.doors[2])
  update(door3, record.doors[3])
  update(door4, record.doors[4])

  address.dataset.original = record.address.configured
  datetime.dataset.original = record.datetime.configured

  return row
}

function combobox(div) {
  const input = div.querySelector('input')
  const list = div.querySelector('ul')
  const now = new Date()
  const options = new Set(
    [...timezones.entries()]
      .map(([tz, f]) => {
        return f(now, tz)
      })
      .sort(),
  )
  const cb = new Combobox(input, list)

  cb.initialise(options)

  return cb
}
