import { schema } from './schema.js'

export const DB = {
  interfaces: new Map(),
  controllers: new Map(),
  doors: new Map(),
  cards: new Map(),
  groups: new Map(),

  tables: {
    events: {
      first: null,
      last: null,
      events: new Map()
    },

    logs: {
      first: null,
      last: null,
      logs: new Map()
    }
  },

  updated: function (tag, recordset) {
    if (recordset) {
      switch (tag) {
        case 'objects':
          recordset.forEach(o => object(o))
          break
      }
    }
  },

  delete: function (tag, oid) {
    switch (tag) {
      case 'controllers':
        if (oid && this.controllers.has(oid)) {
          const record = this.controllers.get(oid)

          record.mark = 0
          record.status = 'deleted'
        }
        break

      case 'doors':
        if (oid && this.doors.has(oid)) {
          const record = this.doors.get(oid)

          record.mark = 0
          record.status = 'deleted'
        }
        break

      case 'cards':
        if (oid && this.cards.has(oid)) {
          const record = this.cards.get(oid)

          record.mark = 0
          record.status = 'deleted'
        }
        break

      case 'groups':
        if (oid && this.groups.has(oid)) {
          const record = this.groups.get(oid)

          record.mark = 0
          record.status = 'deleted'
        }
        break
    }
  },

  refreshed: function (tag) {
    mark(tag)
    sweep(tag)
  },

  events: function () {
    return this.tables.events.events
  },

  firstEvent: function () {
    return this.tables.events.first
  },

  lastEvent: function () {
    return this.tables.events.last
  },

  logs: function () {
    return this.tables.logs.logs
  },

  firstLog: function () {
    return this.tables.logs.first
  },

  lastLog: function () {
    return this.tables.logs.last
  }
}

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
      mark: 0
    })
  }

  const v = DB.interfaces.get(base)

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
      name: '',
      deviceID: '',
      address: { address: '', configured: '', status: 'unknown' },
      datetime: { datetime: '', expected: '', status: 'unknown' },
      cards: { cards: '', status: 'unknown' },
      events: { events: '', status: 'unknown' },
      doors: { 1: '', 2: '', 3: '', 4: '' },
      status: o.value,
      mark: 0
    })
  }

  const v = DB.controllers.get(base)

  switch (oid) {
    case `${base}${schema.controllers.status}`:
      v.status = o.value
      break

    case `${base}${schema.controllers.created}`:
      v.created = o.value
      break

    case `${base}${schema.controllers.name}`:
      v.name = o.value
      break

    case `${base}${schema.controllers.deviceID}`:
      v.deviceID = o.value
      break

    case `${base}${schema.controllers.address}`:
      v.address.address = o.value
      break

    case `${base}${schema.controllers.addressConfigured}`:
      v.address.configured = o.value
      break

    case `${base}${schema.controllers.addressStatus}`:
      v.address.status = o.value
      break

    case `${base}${schema.controllers.datetime}`:
      v.datetime.datetime = o.value
      break

    case `${base}${schema.controllers.expected}`:
      v.datetime.expected = o.value
      break

    case `${base}${schema.controllers.datetimeStatus}`:
      v.datetime.status = o.value
      break

    case `${base}${schema.controllers.cards}`:
      v.cards.cards = o.value
      break

    case `${base}${schema.controllers.cardsStatus}`:
      v.cards.status = o.value
      break

    case `${base}${schema.controllers.events}`:
      v.events.events = o.value
      break

    case `${base}${schema.controllers.eventsStatus}`:
      v.events.status = o.value
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
      status: o.value,
      mark: 0
    })
  }

  const v = DB.doors.get(base)

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
      name: '',
      number: '',
      from: '',
      to: '',
      groups: new Map(),
      status: o.value,
      mark: 0
    })
  }

  const v = DB.cards.get(base)

  switch (oid) {
    case `${base}${schema.cards.status}`:
      v.status = o.value
      break

    case `${base}${schema.cards.created}`:
      v.created = o.value
      break

    case `${base}${schema.cards.name}`:
      v.name = o.value
      break

    case `${base}${schema.cards.card}`:
      v.number = o.value
      break

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
      name: '',
      doors: new Map(),
      status: o.value,
      mark: 0
    })
  }

  const v = DB.groups.get(base)

  switch (oid) {
    case `${base}${schema.groups.status}`:
      v.status = o.value
      break

    case `${base}${schema.groups.created}`:
      v.created = o.value
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
      mark: 0
    })
  }

  const v = DB.tables.events.events.get(base)

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
      mark: 0
    })
  }

  const v = DB.tables.logs.logs.get(base)

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

function mark (tag) {
  DB.controllers.forEach(v => {
    v.mark += 1
  })

  DB.doors.forEach(v => {
    v.mark += 1
  })

  DB.cards.forEach(v => {
    v.mark += 1
  })

  DB.groups.forEach(v => {
    v.mark += 1
  })

  DB.events().forEach(v => {
    v.mark += 1
  })

  DB.logs().forEach(v => {
    v.mark += 1
  })
}

function sweep (tag) {
  DB.controllers.forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.controllers.delete(k)
    }
  })

  DB.doors.forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.doors.delete(k)
    }
  })

  DB.cards.forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.doors.delete(k)
    }
  })

  DB.groups.forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.groups.delete(k)
    }
  })

  DB.events().forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.events().delete(k)
    }
  })

  DB.logs().forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.logs().delete(k)
    }
  })
}
