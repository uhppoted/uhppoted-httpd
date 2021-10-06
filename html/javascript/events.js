import { deleted } from './tabular.js'
import { DB } from './db.js'

HTMLTableSectionElement.prototype.sort = function (cb) {
  Array
    .prototype
    .slice
    .call(this.rows)
    .sort(cb)
    .forEach((e, i, a) => { this.appendChild(this.removeChild(e)) }, this)
}

export function refreshed () {
  const events = [...DB.events.values()]

  realize(events)

  events.forEach(o => {
    const row = updateFromDB(o.OID, o)
    if (row) {
      if (o.status === 'new') {
        row.classList.add('new')
      } else {
        row.classList.remove('new')
      }
    }
  })

  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  tbody.sort((p, q) => {
    const u = DB.events.get(p.dataset.oid)
    const v = DB.events.get(q.dataset.oid)

    return v.timestamp.localeCompare(u.timestamp)
  })

  DB.refreshed('events')
}

function realize (events) {
  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  events.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (o.status === 'deleted') {
      deleted('events', row)
      return
    }

    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid, record) {
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

    // const commit = row.querySelector('td span.commit')
    // commit.id = uuid + '_commit'
    // commit.dataset.record = uuid
    // commit.dataset.enabled = 'false'

    // const rollback = row.querySelector('td span.rollback')
    // rollback.id = uuid + '_rollback'
    // rollback.dataset.record = uuid
    // rollback.dataset.enabled = 'false'

    const fields = [
      { suffix: 'timestamp', oid: `${oid}.3`, selector: 'td input.timestamp', flag: 'td img.timestamp' },
      { suffix: 'deviceID', oid: `${oid}.1`, selector: 'td input.deviceID', flag: 'td img.deviceID' },
      { suffix: 'device', oid: `${oid}.10`, selector: 'td input.device', flag: 'td img.device' },
      { suffix: 'eventType', oid: `${oid}.4`, selector: 'td input.eventType', flag: 'td img.eventType' },
      { suffix: 'doorid', oid: `${oid}.5`, selector: 'td input.doorid', flag: 'td img.doorid' },
      { suffix: 'door', oid: `${oid}.11`, selector: 'td input.door', flag: 'td img.door' },
      { suffix: 'direction', oid: `${oid}.6`, selector: 'td input.direction', flag: 'td img.direction' },
      { suffix: 'cardno', oid: `${oid}.7`, selector: 'td input.cardno', flag: 'td img.cardno' },
      { suffix: 'card', oid: `${oid}.12`, selector: 'td input.card', flag: 'td img.card' },
      { suffix: 'access', oid: `${oid}.8`, selector: 'td input.access', flag: 'td img.access' },
      { suffix: 'reason', oid: `${oid}.9`, selector: 'td input.reason', flag: 'td img.reason' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)
      // const flag = row.querySelector(f.flag)

      if (field) {
        field.id = uuid + '-' + f.suffix
        field.value = ''
        field.dataset.oid = f.oid
        field.dataset.record = uuid
        field.dataset.original = ''
        field.dataset.value = ''

        // flag.id = 'F' + f.oid
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

  const timestamp = row.querySelector(`[data-oid="${oid}.3"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}.1"]`)
  const device = row.querySelector(`[data-oid="${oid}.10"]`)
  const eventType = row.querySelector(`[data-oid="${oid}.4"]`)
  const doorid = row.querySelector(`[data-oid="${oid}.5"]`)
  const door = row.querySelector(`[data-oid="${oid}.11"]`)
  const direction = row.querySelector(`[data-oid="${oid}.6"]`)
  const cardno = row.querySelector(`[data-oid="${oid}.7"]`)
  const card = row.querySelector(`[data-oid="${oid}.12"]`)
  const access = row.querySelector(`[data-oid="${oid}.8"]`)
  const reason = row.querySelector(`[data-oid="${oid}.9"]`)

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

function update (element, value, status) {
  if (element && value !== undefined) {
    element.value = value.toString()
  }
}
