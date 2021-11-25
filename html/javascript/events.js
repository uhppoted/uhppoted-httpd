import { deleted } from './tabular.js'
import { DB } from './db.js'
import { schema } from './schema.js'

export function refreshed () {
  const events = [...DB.events().values()].sort((p, q) => q.timestamp.localeCompare(p.timestamp))
  const pagesize = 5

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
    const table = document.querySelector('#events table')
    const tbody = table.tBodies[0]

    tbody.sort((p, q) => {
      const u = DB.events().get(p.dataset.oid)
      const v = DB.events().get(q.dataset.oid)

      return v.timestamp.localeCompare(u.timestamp)
    })
  }

  // hides/shows the 'more' button
  const h = function () {
    const table = document.querySelector('#events table')
    const tfoot = table.tFoot
    const last = DB.lastEvent()

    if (last && DB.events().has(last)) {
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
    for (let ix = 0; ix < events.length; ix += pagesize) {
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
    .then(() => DB.refreshed('events'))
    .catch(err => console.error(err))
}

function realize (events) {
  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  events.forEach(o => {
    let row = tbody.querySelector(`tr[data-oid="${o.OID}"]`)

    if (o.status === 'deleted') {
      deleted('events', row)
      return
    }

    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid) {
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
      commit.dataset.enabled = 'false'
    }

    const rollback = row.querySelector('td span.rollback')
    if (rollback) {
      rollback.id = uuid + '_rollback'
      rollback.dataset.record = uuid
      rollback.dataset.enabled = 'false'
    }

    const fields = [
      { suffix: 'timestamp', oid: `${oid}${schema.events.timestamp}`, selector: 'td input.timestamp', flag: 'td img.timestamp' },
      { suffix: 'deviceID', oid: `${oid}${schema.events.deviceID}`, selector: 'td input.deviceID', flag: 'td img.deviceID' },
      { suffix: 'device', oid: `${oid}${schema.events.deviceName}`, selector: 'td input.device', flag: 'td img.device' },
      { suffix: 'eventType', oid: `${oid}${schema.events.type}`, selector: 'td input.eventType', flag: 'td img.eventType' },
      { suffix: 'doorid', oid: `${oid}${schema.events.door}`, selector: 'td input.doorid', flag: 'td img.doorid' },
      { suffix: 'door', oid: `${oid}${schema.events.doorName}`, selector: 'td input.door', flag: 'td img.door' },
      { suffix: 'direction', oid: `${oid}${schema.events.direction}`, selector: 'td input.direction', flag: 'td img.direction' },
      { suffix: 'cardno', oid: `${oid}${schema.events.card}`, selector: 'td input.cardno', flag: 'td img.cardno' },
      { suffix: 'card', oid: `${oid}${schema.events.cardName}`, selector: 'td input.card', flag: 'td img.card' },
      { suffix: 'access', oid: `${oid}${schema.events.granted}`, selector: 'td input.access', flag: 'td img.access' },
      { suffix: 'reason', oid: `${oid}${schema.events.reason}`, selector: 'td input.reason', flag: 'td img.reason' }
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

        if (flag) {
          flag.id = 'F' + f.oid
        }
      } else {
        console.error(f)
      }
    })

    return row
  }
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#events tr[data-oid='" + oid + "']")

  if (record.status === 'deleted' || !row) {
    return
  }

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
  update(device, record.deviceName.toLowerCase())
  update(eventType, record.eventType)
  update(doorid, record.door)
  update(door, record.doorName.toLowerCase())
  update(direction, record.direction)
  update(cardno, record.card)
  update(card, record.cardName.toLowerCase())
  update(access, record.granted === 'true' ? 'granted' : (record.granted === 'false' ? 'denied' : ''))
  update(reason, record.reason)

  return row
}

function update (element, value) {
  if (element && value !== undefined) {
    element.value = value.toString()
  }
}
