import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'

const pagesize = 5

export function refreshed () {
  refreshControllers()
  refreshEvents()
  refreshLogs()
}

function refreshControllers () {
  const list = [...DB.controllers.values()]
    .filter(c => alive(c))
    .sort((p, q) => p.created.localeCompare(q.created))

  realizeControllers(list)

  list.forEach(o => {
    updateController(o.OID, o)
  })
}

function refreshEvents () {
  const events = [...DB.events().values()]
    .filter(e => alive(e))
    .sort((p, q) => q.timestamp.localeCompare(p.timestamp))
    .slice(0, pagesize)

  realizeEvents(events)

  events.forEach(o => {
    updateEvent(o.OID, o)
  })

  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  tbody.sort((p, q) => {
    const u = DB.events().get(p.dataset.oid)
    const v = DB.events().get(q.dataset.oid)

    return v.timestamp.localeCompare(u.timestamp)
  })
}

function refreshLogs () {
  const logs = [...DB.logs().values()]
    .filter(l => alive(l))
    .sort((p, q) => q.timestamp.localeCompare(p.timestamp))
    .slice(0, pagesize)

  realizeLogs(logs)

  logs.forEach(o => {
    updateLog(o.OID, o)
  })

  // sorts the table rows by 'timestamp'
    const table = document.querySelector('#logs table')
    const tbody = table.tBodies[0]

    tbody.sort((p, q) => {
      const u = DB.logs().get(p.dataset.oid)
      const v = DB.logs().get(q.dataset.oid)

      return v.timestamp.localeCompare(u.timestamp)
    })
}

function realizeControllers (controllers) {
  const table = document.querySelector('#controllers table')
  const tbody = table.tBodies[0]

  trim('controllers', controllers, tbody.querySelectorAll('tr.controller'))

  controllers.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")

    if (!row) {
      row = addController(o.OID, o)
    }
  })
}

function addController (oid, record) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('controllers').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#controller')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('controller')
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

function updateController (oid, record) {
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

function realizeEvents (events) {
  const table = document.querySelector('#events table')
  const tbody = table.tBodies[0]

  trim('events', events, tbody.querySelectorAll('tr.event'))

  events.forEach(o => {
    let row = tbody.querySelector(`tr[data-oid="${o.OID}"]`)
    if (!row) {
      row = addEvent(o.OID, o)
    }
  })
}

function addEvent (oid) {
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

    const fields = [
      { suffix: 'timestamp', oid: `${oid}${schema.events.timestamp}`, selector: 'td input.timestamp', flag: 'td img.timestamp' },
      { suffix: 'device', oid: `${oid}${schema.events.deviceName}`, selector: 'td input.device', flag: 'td img.device' },
      { suffix: 'eventType', oid: `${oid}${schema.events.type}`, selector: 'td input.eventType', flag: 'td img.eventType' },
      { suffix: 'door', oid: `${oid}${schema.events.doorName}`, selector: 'td input.door', flag: 'td img.door' },
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

function updateEvent (oid, record) {
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

function realizeLogs (logs) {
  const table = document.querySelector('#logs table')
  const tbody = table.tBodies[0]

  trim('logs', logs, tbody.querySelectorAll('tr.entry'))

  logs.forEach(o => {
    let row = tbody.querySelector(`tr[data-oid='${o.OID}']`)
    if (!row) {
      row = addLog(o.OID, o)
    }
  })
}

function addLog (oid) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('logs').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('#entry')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('entry')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const fields = [
      { suffix: 'timestamp', oid: `${oid}${schema.logs.timestamp}`, selector: 'td input.timestamp' },
      { suffix: 'uid', oid: `${oid}${schema.logs.uid}`, selector: 'td input.uid' },
      { suffix: 'details', oid: `${oid}${schema.logs.details}`, selector: 'td input.details' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)
      const flag = row.querySelector(`td img.${f.suffix}`)

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

function updateLog (oid, record) {
  const row = document.querySelector("div#logs tr[data-oid='" + oid + "']")

  const timestamp = row.querySelector(`[data-oid="${oid}${schema.logs.timestamp}"]`)
  const uid = row.querySelector(`[data-oid="${oid}${schema.logs.uid}"]`)
  const details = row.querySelector(`[data-oid="${oid}${schema.logs.details}"]`)

  row.dataset.status = record.status

  update(timestamp, format(record.timestamp))
  update(uid, record.uid)
  update(details, record.item.details)

  return row
}

function format (timestamp) {
  const dt = Date.parse(timestamp)
  const fmt = function (v) {
    return v < 10 ? '0' + v.toString() : v.toString()
  }

  if (!isNaN(dt)) {
    const date = new Date(dt)
    const year = date.getFullYear()
    const month = fmt(date.getMonth() + 1)
    const day = fmt(date.getDate())
    const hour = fmt(date.getHours())
    const minute = fmt(date.getMinutes())
    const second = fmt(date.getSeconds())

    return `${year}-${month}-${day} ${hour}:${minute}:${second}`
  }

  return ''
}
