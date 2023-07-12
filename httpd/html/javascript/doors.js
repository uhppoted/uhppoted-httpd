import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { Combobox } from './combobox.js'

export function refreshed () {
  const doors = [...DB.doors.values()]
    .filter(d => alive(d))
    .sort((p, q) => p.created.localeCompare(q.created))

  realize(doors)

  doors.forEach(d => {
    const row = updateFromDB(d.OID, d)
    if (row) {
      if (d.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })
}

export function deletable (row) {
  const name = row.querySelector('td input.name')
  const re = /^\s*$/

  if (name && name.dataset.oid !== '' && re.test(name.dataset.value)) {
    return true
  }

  return false
}

function realize (doors) {
  const table = document.querySelector('#doors table')
  const tbody = table.tBodies[0]

  // ... rows
  trim('doors', doors, tbody.querySelectorAll('tr.door'))

  doors.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('doors').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#door')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('door')
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
      { suffix: 'name', oid: `${oid}.1`, selector: 'td input.name' },
      { suffix: 'controller', oid: `${oid}.0.4.2`, selector: 'td input.controller' },
      { suffix: 'deviceID', oid: `${oid}.0.4.3`, selector: 'td input.deviceID' },
      { suffix: 'doorID', oid: `${oid}.0.4.4`, selector: 'td input.doorID' },
      { suffix: 'delay', oid: `${oid}.2`, selector: 'td input.delay' },
      { suffix: 'mode', oid: `${oid}.3`, selector: 'td input.mode' },
      { suffix: 'keypad', oid: `${oid}.4`, selector: 'td label.keypad input' }
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

    combobox(row.querySelector('td.combobox'))

    return row
  }
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#doors tr[data-oid='" + oid + "']")

  const name = row.querySelector(`[data-oid="${oid}.1"]`)
  const controller = row.querySelector(`[data-oid="${oid}.0.4.2"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}.0.4.3"]`)
  const door = row.querySelector(`[data-oid="${oid}.0.4.4"]`)
  const delay = row.querySelector(`[data-oid="${oid}.2"]`)
  const mode = row.querySelector(`[data-oid="${oid}.3"]`)
  const keypad = row.querySelector(`[data-oid="${oid}.4"]`)

  row.dataset.status = record.status

  const d = record.delay.status === 'uncertain' ? record.delay.configured : record.delay.delay
  const m = record.mode.status === 'uncertain' ? record.mode.configured : record.mode.mode
  const c = lookup(record)

  update(name, record.name)
  update(controller, c.name)
  update(deviceID, c.deviceID)
  update(door, c.door)
  update(delay, d, record.delay.status)
  update(mode, m, record.mode.status)
  update(keypad, record.keypad)

  // ... set tooltips for error'd values
  { const tooltip = row.querySelector(`[data-oid="${oid}.2"] + div.tooltip-content`)

    if (tooltip) {
      const p = tooltip.querySelector('p')
      const err = record.delay.err && record.delay.err !== '' ? record.delay.err : ''
      const enabled = !!(record.delay.err && record.delay.err !== '')

      p.innerHTML = err

      if (enabled) {
        tooltip.classList.add('enabled')
      } else {
        tooltip.classList.remove('enabled')
      }
    }
  }

  { const tooltip = row.querySelector(`[data-oid="${oid}.3"] + ul + div`)

    if (tooltip) {
      const p = tooltip.querySelector('p')
      const err = record.mode.err && record.mode.err !== '' ? record.mode.err : ''
      const enabled = !!(record.mode.err && record.mode.err !== '')

      p.innerHTML = err

      if (enabled) {
        tooltip.classList.add('enabled')
      } else {
        tooltip.classList.remove('enabled')
      }
    }
  }

  return row
}

function lookup (record) {
  const oid = record.OID

  const object = {
    name: '',
    deviceID: '',
    door: ''
  }

  const controller = [...DB.controllers.values()].find(c => {
    for (const d of [1, 2, 3, 4]) {
      if (c.doors[d] === oid) {
        return true
      }
    }

    return false
  })

  if (controller) {
    object.name = controller.name
    object.deviceID = controller.deviceID

    for (const d of [1, 2, 3, 4]) {
      if (controller.doors[d] === oid) {
        object.door = d
      }
    }
  }

  return object
}

function combobox (div) {
  const input = div.querySelector('input')
  const list = div.querySelector('ul')
  const options = new Set(['controlled', 'normally open', 'normally closed'])
  const cb = new Combobox(input, list)

  cb.initialise(options)

  return cb
}
