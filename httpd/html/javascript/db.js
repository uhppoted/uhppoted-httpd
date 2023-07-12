import { schema } from './schema.js'

class DBC {
  constructor () {
    this.interfaces = new Map()
    this.controllers = new Map()
    this.doors = new Map()
    this.cards = new Map()
    this.groups = new Map()

    this.init = function () {
      setInterval(this.sweep, 15000)
    }

    this.tables = {
      events: {
        status: 'unknown',
        first: null,
        last: null,
        events: new Map()
      },

      logs: {
        first: null,
        last: null,
        logs: new Map()
      },

      users: {
        users: new Map()
      }
    }

    this.get = function (oid) {
      return [null, false]
    }

    this.updated = function (tag, recordset) {
      if (recordset) {
        switch (tag) {
          case 'interfaces':
          case 'controllers':
          case 'doors':
          case 'cards':
          case 'groups':
          case 'events':
          case 'logs':
          case 'users':
            recordset.forEach(o => object(o))
            break
        }
      }
    }

    this.delete = function (tag, oid) {
      if (oid) {
        switch (tag) {
          case 'interfaces':
            this.interfaces.delete(oid)
            break

          case 'controllers':
            this.controllers.delete(oid)
            break

          case 'doors':
            this.doors.delete(oid)
            break

          case 'cards':
            this.cards.delete(oid)
            break

          case 'groups':
            this.groups.delete(oid)
            break

          case 'users':
            this.tables.users.users.delete(oid)
            break
        }
      }
    }

    this.events = function () {
      return this.tables.events.events
    }

    this.firstEvent = function () {
      return this.tables.events.first
    }

    this.lastEvent = function () {
      return this.tables.events.last
    }

    this.logs = function () {
      return this.tables.logs.logs
    }

    this.firstLog = function () {
      return this.tables.logs.first
    }

    this.lastLog = function () {
      return this.tables.logs.last
    }

    this.users = function () {
      return this.tables.users.users
    }

    this.sweep = function () {
      sweep()
    }

    this.init()
  }
}

export const DB = new DBC()

function object (o) {
  const oid = o.OID

  if (oid.startsWith(schema.interfaces.base)) {
    interfaces(o)
  } else if (oid.startsWith(schema.controllers.base)) {
    controllers(o)
  } else if (oid.startsWith(schema.doors.base)) {
    doors(o)
  } else if (oid.startsWith(schema.cards.base)) {
    cards(o)
  } else if (oid.startsWith(schema.groups.base)) {
    groups(o)
  } else if (oid.startsWith(schema.events.base)) {
    events(o)
  } else if (oid.startsWith(schema.logs.base)) {
    logs(o)
  } else if (oid.startsWith(schema.users.base)) {
    users(o)
  }
}

function interfaces (o) {
  const oid = o.OID
  const match = oid.match(schema.interfaces.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.interfaces.has(base)) {
    DB.interfaces.set(base, {
      OID: oid,
      type: 'LAN',
      name: 'LAN',
      bind: '',
      broadcast: '',
      listen: '',
      status: '',
      touched: new Date()
    })
  }

  const v = DB.interfaces.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.interfaces.status}`:
      v.status = o.value
      break

    case `${base}${schema.interfaces.type}`:
      v.type = o.value
      break

    case `${base}${schema.interfaces.name}`:
      v.name = o.value
      break

    case `${base}${schema.interfaces.bind}`:
      v.bind = o.value
      break

    case `${base}${schema.interfaces.broadcast}`:
      v.broadcast = o.value
      break

    case `${base}${schema.interfaces.listen}`:
      v.listen = o.value
      break
  }
}

function controllers (o) {
  const oid = o.OID
  const match = oid.match(schema.controllers.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.controllers.has(base)) {
    DB.controllers.set(base, {
      OID: oid,
      created: '',
      deleted: '',
      name: '',
      deviceID: '',
      address: { address: '', configured: '', status: 'unknown' },
      datetime: { datetime: '', configured: '', status: 'unknown' },
      interlock: '',
      cards: { cards: '', status: 'unknown' },
      events: { events: '', status: 'unknown' },
      doors: { 1: '', 2: '', 3: '', 4: '' },
      status: o.value,
      touched: new Date()
    })
  }

  const v = DB.controllers.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.controllers.status}`:
      v.status = o.value
      break

    case `${base}${schema.controllers.created}`:
      v.created = o.value
      break

    case `${base}${schema.controllers.deleted}`:
      v.deleted = o.value
      break

    case `${base}${schema.controllers.name}`:
      v.name = o.value
      break

    case `${base}${schema.controllers.deviceID}`:
      v.deviceID = o.value
      break

    case `${base}${schema.controllers.endpoint.status}`:
      v.address.status = o.value
      break

    case `${base}${schema.controllers.endpoint.address}`:
      v.address.address = o.value
      break

    case `${base}${schema.controllers.endpoint.configured}`:
      v.address.configured = o.value
      break

    case `${base}${schema.controllers.datetime.status}`:
      v.datetime.status = o.value
      break

    case `${base}${schema.controllers.datetime.current}`:
      v.datetime.datetime = o.value
      break

    case `${base}${schema.controllers.datetime.configured}`:
      v.datetime.configured = o.value
      break

    case `${base}${schema.controllers.interlock}`:
      v.interlock = o.value
      break

    case `${base}${schema.controllers.cards.status}`:
      v.cards.status = o.value
      break

    case `${base}${schema.controllers.cards.count}`:
      v.cards.cards = o.value
      break

    case `${base}${schema.controllers.events.status}`:
      v.events.status = o.value
      break

    case `${base}${schema.controllers.events.first}`:
      v.events.first = o.value
      break

    case `${base}${schema.controllers.events.last}`:
      v.events.last = o.value
      break

    case `${base}${schema.controllers.events.current}`:
      v.events.current = o.value
      break

    case `${base}${schema.controllers.door1}`:
      v.doors[1] = o.value
      break

    case `${base}${schema.controllers.door2}`:
      v.doors[2] = o.value
      break

    case `${base}${schema.controllers.door3}`:
      v.doors[3] = o.value
      break

    case `${base}${schema.controllers.door4}`:
      v.doors[4] = o.value
      break
  }
}

function doors (o) {
  const oid = o.OID
  const match = oid.match(schema.doors.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.doors.has(base)) {
    DB.doors.set(base, {
      OID: oid,
      created: '',
      deleted: '',
      name: '',
      delay: { delay: '', configured: '', status: 'unknown', err: '' },
      mode: { mode: '', configured: '', status: 'unknown', err: '' },
      keypad: false,
      status: o.value,
      touched: new Date()
    })
  }

  const v = DB.doors.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.doors.status}`:
      v.status = o.value
      break

    case `${base}${schema.doors.created}`:
      v.created = o.value
      break

    case `${base}${schema.doors.deleted}`:
      v.deleted = o.value
      break

    case `${base}.1`:
      v.name = o.value
      break

    case `${base}.2`:
      v.delay.delay = o.value
      break

    case `${base}.2.1`:
      v.delay.status = o.value
      break

    case `${base}.2.2`:
      v.delay.configured = o.value
      break

    case `${base}.2.3`:
      v.delay.err = o.value
      break

    case `${base}.3`:
      v.mode.mode = o.value
      break

    case `${base}.3.1`:
      v.mode.status = o.value
      break

    case `${base}.3.2`:
      v.mode.configured = o.value
      break

    case `${base}.3.3`:
      v.mode.err = o.value
      break

    case `${base}.4`:
      v.keypad = o.value === 'true'
      break
  }
}

function cards (o) {
  const oid = o.OID
  const match = oid.match(schema.cards.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.cards.has(base)) {
    DB.cards.set(base, {
      OID: oid,
      created: '',
      deleted: '',
      name: '',
      number: '',
      // {{if .WithPIN}}
      PIN: '',
      // {{end}}
      from: '',
      to: '',
      groups: new Map(),
      status: o.value,
      touched: new Date()
    })
  }

  const v = DB.cards.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.cards.status}`:
      v.status = o.value
      break

    case `${base}${schema.cards.created}`:
      v.created = o.value
      break

    case `${base}${schema.cards.deleted}`:
      v.deleted = o.value
      break

    case `${base}${schema.cards.name}`:
      v.name = o.value
      break

    case `${base}${schema.cards.card}`:
      v.number = o.value
      break

    // {{if .WithPIN}}
    case `${base}${schema.cards.PIN}`:
      v.PIN = o.value
      break
      // {{end}}

    case `${base}${schema.cards.from}`:
      v.from = o.value
      break

    case `${base}${schema.cards.to}`:
      v.to = o.value
      break

    default: {
      const m = oid.match(schema.cards.groups)
      if (m && m.length > 2) {
        const suboid = m[1]
        const suffix = m[2]

        if (!v.groups.has(suboid)) {
          v.groups.set(suboid, { group: '', member: false })
        }

        const group = v.groups.get(suboid)

        if (!suffix) {
          group.member = o.value === 'true'
        } else if (suffix === '.1') {
          group.group = o.value
        }
      }
    }
  }
}

function groups (o) {
  const oid = o.OID
  const match = oid.match(schema.groups.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.groups.has(base)) {
    DB.groups.set(base, {
      OID: oid,
      created: '',
      deleted: '',
      name: '',
      doors: new Map(),
      status: o.value,
      touched: new Date()
    })
  }

  const v = DB.groups.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.groups.status}`:
      v.status = o.value
      break

    case `${base}${schema.groups.created}`:
      v.created = o.value
      break

    case `${base}${schema.groups.deleted}`:
      v.deleted = o.value
      break

    case `${base}${schema.groups.name}`:
      v.name = o.value
      break

    default: {
      const m = oid.match(schema.groups.doors)
      if (m && m.length > 2) {
        const suboid = m[1]
        const suffix = m[2]

        if (!v.doors.has(suboid)) {
          v.doors.set(suboid, { door: '', allowed: false })
        }

        const door = v.doors.get(suboid)

        if (!suffix) {
          door.allowed = o.value === 'true'
        } else if (suffix === '.1') {
          door.door = o.value
        }
      }
    }
  }
}

function events (o) {
  const oid = o.OID

  if (oid === `${schema.events.base}${schema.events.status}`) {
    DB.tables.events.status = o.value
    return
  }

  if (oid === `${schema.events.base}${schema.events.first}`) {
    DB.tables.events.first = o.value
    return
  }

  if (oid === `${schema.events.base}${schema.events.first}`) {
    DB.tables.events.last = o.value
    return
  }

  const match = oid.match(schema.events.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.tables.events.events.has(base)) {
    DB.tables.events.events.set(base, {
      OID: oid,
      timestamp: '',
      deviceID: '',
      index: 0,
      eventType: '',
      door: '',
      direction: '',
      card: '',
      granted: '',
      reason: '',
      deviceName: '',
      doorName: '',
      cardName: '',
      touched: new Date()
    })
  }

  const v = DB.tables.events.events.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.events.timestamp}`:
      v.timestamp = o.value
      break

    case `${base}${schema.events.deviceID}`:
      v.deviceID = o.value
      break

    case `${base}${schema.events.index}`:
      v.index = parseInt(o.value)
      break

    case `${base}${schema.events.type}`:
      v.eventType = o.value
      break

    case `${base}${schema.events.door}`:
      v.door = o.value
      break

    case `${base}${schema.events.direction}`:
      v.direction = o.value
      break

    case `${base}${schema.events.card}`:
      v.card = o.value
      break

    case `${base}${schema.events.granted}`:
      v.granted = o.value
      break

    case `${base}${schema.events.reason}`:
      v.reason = o.value
      break

    case `${base}${schema.events.deviceName}`:
      v.deviceName = o.value
      break

    case `${base}${schema.events.doorName}`:
      v.doorName = o.value
      break

    case `${base}${schema.events.cardName}`:
      v.cardName = o.value
      break
  }
}

function logs (o) {
  const oid = o.OID

  if (oid === `${schema.logs.base}${schema.logs.first}`) {
    DB.tables.logs.first = o.value
    return
  }

  if (oid === `${schema.logs.base}${schema.logs.last}`) {
    DB.tables.logs.last = o.value
    return
  }

  const match = oid.match(schema.logs.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.tables.logs.logs.has(base)) {
    DB.tables.logs.logs.set(base, {
      OID: oid,
      timestamp: '',
      uid: '',
      item: {
        type: '',
        ID: '',
        name: '',
        field: ''
      },
      details: '',
      touched: new Date()
    })
  }

  const v = DB.tables.logs.logs.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.logs.timestamp}`:
      v.timestamp = o.value
      break

    case `${base}${schema.logs.uid}`:
      v.uid = o.value
      break

    case `${base}${schema.logs.item}`:
      v.item.type = o.value
      break

    case `${base}${schema.logs.itemID}`:
      v.item.ID = o.value
      break

    case `${base}${schema.logs.itemName}`:
      v.item.name = o.value
      break

    case `${base}${schema.logs.field}`:
      v.item.field = o.value
      break

    case `${base}${schema.logs.details}`:
      v.item.details = o.value
      break
  }
}

function users (o) {
  const oid = o.OID

  const match = oid.match(schema.users.regex)

  if (!match || match.length < 2) {
    return
  }

  const base = match[1]

  if (!DB.tables.users.users.has(base)) {
    DB.tables.users.users.set(base, {
      OID: oid,
      name: '',
      uid: '',
      role: '',
      password: '',
      otp: '',
      locked: '',
      details: '',
      created: '',
      deleted: '',
      status: o.value,
      touched: new Date()
    })
  }

  const v = DB.tables.users.users.get(base)

  v.touched = new Date()

  switch (oid) {
    case `${base}${schema.users.status}`:
      v.status = o.value
      break

    case `${base}${schema.users.created}`:
      v.created = o.value
      break

    case `${base}${schema.users.deleted}`:
      v.deleted = o.value
      break

    case `${base}${schema.users.name}`:
      v.name = o.value
      break

    case `${base}${schema.users.uid}`:
      v.uid = o.value
      break

    case `${base}${schema.users.role}`:
      v.role = o.value
      break

    case `${base}${schema.users.password}`:
      v.password = o.value
      break

    case `${base}${schema.users.otp}`:
      v.otp = o.value
      break

    case `${base}${schema.users.locked}`:
      v.locked = o.value
      break
  }
}

function sweep () {
  const tables = [DB.interfaces, DB.controllers, DB.doors, DB.cards, DB.groups]
  const now = new Date()
  const sweepable = 5 * 60 * 1000 // 5 minutes

  tables.forEach(t => {
    t.forEach((v, k) => {
      const dt = now - v.touched
      if (!Number.isNaN(dt) && dt > sweepable && !alive(v)) {
        t.delete(k)
      }
    })
  })
}

export function alive (object) {
  if (object) {
    return !(object.deleted && object.deleted !== '')
  }

  return false
}
