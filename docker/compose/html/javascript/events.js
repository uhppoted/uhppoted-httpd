import { trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'
import { loaded } from './uhppoted.js'

const pagesize = 5

export function refreshed() {
  const events = [...DB.events().values()]
    .filter((e) => alive(e))
    .filter((e) => e.timestamp !== '')
    .sort((p, q) => q.timestamp.localeCompare(p.timestamp))

  realize(events)

  // renders a 'page size' of events
  const f = function (offset) {
    let ix = offset
    let count = 0
    while (count < pagesize && ix < events.length) {
      const o = events[ix]
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
    const focused = document.activeElement

    if (!focused || focused.nodeName !== 'INPUT') {
      const table = document.querySelector('#events table')
      const tbody = table.tBodies[0]

      tbody.sort((p, q) => {
        const u = DB.events().get(p.dataset.oid)
        const v = DB.events().get(q.dataset.oid)

        return v.timestamp.localeCompare(u.timestamp)
      })
    }
  }

  // hides/shows the 'more' button
  const h = function () {
    const table = document.querySelector('#events table')
    const tfoot = table.tFoot
    const first = DB.firstEvent()

    if (first && DB.events().has(first)) {
      tfoot.classList.add('hidden')
    } else {
      tfoot.classList.remove('hidden')
    }
  }

  // initialises the rows asynchronously in small'ish chunks
  const chunk = (offset) =>
    new Promise((resolve) => {
      f(offset)
      resolve(true)
    })

  async function* render() {
    for (let ix = 0; ix < events.length; ix += pagesize) {
      yield chunk(ix).then(() => ix)
    }
  }

  ;(async function loop() {
    for await (const _ of render()) {
      // empty
    }
  })()
    .then(() => g())
    .then(() => h())
    .catch((err) => console.error(err))
    .finally(() => {
      loaded()
    })
}

function realize(events) {
  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  trim('events', events, tbody.querySelectorAll('tr.event'))

  events.forEach((o) => {
    let row = tbody.querySelector(`tr[data-oid="${o.OID}"]`)
    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add(oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('events').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#event')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('event')
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
      {
        suffix: 'timestamp',
        oid: `${oid}${schema.events.timestamp}`,
        selector: 'td input.timestamp',
      },
      {
        suffix: 'deviceID',
        oid: `${oid}${schema.events.deviceID}`,
        selector: 'td input.deviceID',
      },
      {
        suffix: 'device',
        oid: `${oid}${schema.events.deviceName}`,
        selector: 'td input.device',
      },
      {
        suffix: 'eventType',
        oid: `${oid}${schema.events.type}`,
        selector: 'td input.eventType',
      },
      {
        suffix: 'doorid',
        oid: `${oid}${schema.events.door}`,
        selector: 'td input.doorid',
      },
      {
        suffix: 'door',
        oid: `${oid}${schema.events.doorName}`,
        selector: 'td input.door',
      },
      {
        suffix: 'direction',
        oid: `${oid}${schema.events.direction}`,
        selector: 'td input.direction',
      },
      {
        suffix: 'cardno',
        oid: `${oid}${schema.events.card}`,
        selector: 'td input.cardno',
      },
      {
        suffix: 'card',
        oid: `${oid}${schema.events.cardName}`,
        selector: 'td input.card',
      },
      {
        suffix: 'access',
        oid: `${oid}${schema.events.granted}`,
        selector: 'td input.access',
      },
      {
        suffix: 'reason',
        oid: `${oid}${schema.events.reason}`,
        selector: 'td input.reason',
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
      } else {
        console.error(f)
      }
    })

    return row
  }
}

function updateFromDB(oid, record) {
  const row = document.querySelector("div#events tr[data-oid='" + oid + "']")

  const timestamp = row.querySelector(`[data-oid="${oid}${schema.events.timestamp}"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}${schema.events.deviceID}"]`)
  const device = row.querySelector(`[data-oid="${oid}${schema.events.deviceName}"]`)
  const eventType = row.querySelector(`[data-oid="${oid}${schema.events.type}"]`)
  const doorid = row.querySelector(`[data-oid="${oid}${schema.events.door}"]`)
  const door = row.querySelector(`[data-oid="${oid}${schema.events.doorName}"]`)
  const direction = row.querySelector(`[data-oid="${oid}${schema.events.direction}"]`)
  const cardno = row.querySelector(`[data-oid="${oid}${schema.events.card}"]`)
  const card = row.querySelector(`[data-oid="${oid}${schema.events.cardName}"]`)
  const access = row.querySelector(`[data-oid="${oid}${schema.events.granted}"]`)
  const reason = row.querySelector(`[data-oid="${oid}${schema.events.reason}"]`)

  row.dataset.status = record.status

  update(timestamp, record.timestamp)
  update(deviceID, record.deviceID)
  update(device, record.deviceName)
  update(eventType, record.eventType)
  update(doorid, record.door)
  update(door, record.doorName)
  update(direction, record.direction)
  update(cardno, record.card)
  update(card, record.cardName)
  update(access, record.granted === 'true' ? 'granted' : record.granted === 'false' ? 'denied' : '')
  update(reason, record.reason)

  return row
}

function update(element, value) {
  if (element && value !== undefined) {
    element.value = value.toString()
  }
}
