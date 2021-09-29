import { update, deleted } from './tabular.js'
import { DB } from './db.js'

export function refreshed () {
  const events = [...DB.events.values()] // .sort((p, q) => p.index - q.index)

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

  // DB.refreshed('groups')
}

function realize (events) {
  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  // ... rows

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
    // row.classList.add('new')
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
      { suffix: 'timestamp', oid: `${oid}.3`, selector: 'td input.timestamp', flag: 'td img.timestamp' },
      { suffix: 'deviceID', oid: `${oid}.1`, selector: 'td input.deviceID', flag: 'td img.deviceID' },
      { suffix: 'eventType', oid: `${oid}.4`, selector: 'td input.eventType', flag: 'td img.eventType' },
      { suffix: 'door', oid: `${oid}.5`, selector: 'td input.door', flag: 'td img.door' },
      { suffix: 'direction', oid: `${oid}.6`, selector: 'td input.direction', flag: 'td img.direction' },
      { suffix: 'card', oid: `${oid}.7`, selector: 'td input.card', flag: 'td img.card' },
      { suffix: 'access', oid: `${oid}.8`, selector: 'td input.access', flag: 'td img.access' },
      { suffix: 'reason', oid: `${oid}.9`, selector: 'td input.reason', flag: 'td img.reason' }
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

function updateFromDB (oid, record) {
  const row = document.querySelector("div#events tr[data-oid='" + oid + "']")

  if (record.status === 'deleted' || !row) {
    return
  }

  const timestamp = row.querySelector(`[data-oid="${oid}.3"]`)
  const deviceID = row.querySelector(`[data-oid="${oid}.1"]`)
  const eventType = row.querySelector(`[data-oid="${oid}.4"]`)
  const door = row.querySelector(`[data-oid="${oid}.5"]`)
  const direction = row.querySelector(`[data-oid="${oid}.6"]`)
  const card = row.querySelector(`[data-oid="${oid}.7"]`)
  const access = row.querySelector(`[data-oid="${oid}.8"]`)
  const reason = row.querySelector(`[data-oid="${oid}.9"]`)

  row.dataset.status = record.status

  update(timestamp, record.timestamp)
  update(deviceID, record.deviceID)
  update(eventType, record.eventType)
  update(door, record.door)
  update(direction, record.direction)
  update(card, record.card)
  update(access, record.granted === 'true' ? 'granted' : (record.granted === 'false' ? 'denied' : ''))
  update(reason, record.reason)

  return row
}
