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
    DB.interfaces.set(oid, {
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
    case base + schema.interfaces.status:
      v.status = o.value
      break

    case base + schema.interfaces.type:
      v.type = o.value
      break

    case base + schema.interfaces.name:
      v.name = o.value
      break

    case base + schema.interfaces.bind:
      v.bind = o.value
      break

    case base + schema.interfaces.broadcast:
      v.broadcast = o.value
      break

    case base + schema.interfaces.listen:
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
    DB.controllers.set(oid, {
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
    case base + schema.controllers.status:
      v.status = o.value
      break

    case base + schema.controllers.created:
      v.created = o.value
      break

    case base + '.1':
      v.name = o.value
      break

    case base + '.2':
      v.deviceID = o.value
      break

    case base + '.3':
      v.address.address = o.value
      break

    case base + '.3.1':
      v.address.configured = o.value
      break

    case base + '.3.2':
      v.address.status = o.value
      break

    case base + '.4':
      v.datetime.datetime = o.value
      break

    case base + '.4.1':
      v.datetime.expected = o.value
      break

    case base + '.4.2':
      v.datetime.status = o.value
      break

    case base + '.5':
      v.cards.cards = o.value
      break

    case base + '.5.1':
      v.cards.status = o.value
      break

    case base + '.6':
      v.events.events = o.value
      break

    case base + '.6.1':
      v.events.status = o.value
      break

    case base + '.7':
      v.doors[1] = o.value
      break

    case base + '.8':
      v.doors[2] = o.value
      break

    case base + '.9':
      v.doors[3] = o.value
      break

    case base + '.10':
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
    DB.doors.set(oid, {
      OID: oid,
      created: '',
      controller: '',
      deviceID: '',
      door: '',
      name: '',
      delay: { delay: '', configured: '', status: 'unknown', err: '' },
      mode: { mode: '', configured: '', status: 'unknown', err: '' },
      status: o.value,
      mark: 0
    })
  }

  const v = DB.doors.get(base)

  switch (oid) {
    case base + schema.doors.status:
      v.status = o.value
      break

    case base + schema.doors.created:
      v.created = o.value
      break

    case base + '.0.2.2':
      v.controller = o.value
      break

    case base + '.0.2.3':
      v.deviceID = o.value
      break

    case base + '.0.2.4':
      v.door = o.value
      break

    case base + '.1':
      v.name = o.value
      break

    case base + '.2':
      v.delay.delay = o.value
      break

    case base + '.2.1':
      v.delay.status = o.value
      break

    case base + '.2.2':
      v.delay.configured = o.value
      break

    case base + '.2.3':
      v.delay.err = o.value
      break

    case base + '.3':
      v.mode.mode = o.value
      break

    case base + '.3.1':
      v.mode.status = o.value
      break

    case base + '.3.2':
      v.mode.configured = o.value
      break

    case base + '.3.3':
      v.mode.err = o.value
      break
  }
}

function cards (o) {
  const oid = o.OID

  if (schema.cards.regex.test(oid)) {
    if (DB.cards.has(oid)) {
      const record = DB.cards.get(oid)
      record.status = o.value
      record.mark = 0
      return
    }

    DB.cards.set(oid, {
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

    return
  }

  DB.cards.forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k:
          v.status = o.value
          break

        case k + '.0.1':
          v.created = o.value
          break

        case k + '.1':
          v.name = o.value
          break

        case k + '.2':
          v.number = o.value
          break

        case k + '.3':
          v.from = o.value
          break

        case k + '.4':
          v.to = o.value
          break

        default:
          if (oid.startsWith(k + '.5.')) {
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
  })
}

function groups (o) {
  const oid = o.OID

  if (schema.groups.regex.test(oid)) {
    if (DB.groups.has(oid)) {
      const record = DB.groups.get(oid)
      record.status = o.value
      record.mark = 0
      return
    }

    DB.groups.set(oid, {
      OID: oid,
      created: '',
      name: '',
      doors: new Map(),
      index: 0,
      status: o.value,
      mark: 0
    })

    return
  }

  DB.groups.forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k:
          v.status = o.value
          break

        case k + '.0.1':
          v.created = o.value
          break

        case k + '.1':
          v.name = o.value
          break

        case k + '.3':
          v.index = parseInt(o.value, 10)
          break

        default:
          if (oid.startsWith(k + '.2.')) {
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
  })
}

function events (o) {
  const oid = o.OID

  if (oid === schema.events.base + '.0.1') {
    DB.tables.events.first = o.value
    return
  }

  if (oid === schema.events.base + '.0.2') {
    DB.tables.events.last = o.value
    return
  }

  if (schema.events.regex.test(oid)) {
    if (DB.events().has(oid)) {
      const record = DB.events().get(oid)
      record.status = o.value
      record.mark = 0
      return
    }

    DB.events().set(oid, {
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
      status: o.value,
      mark: 0
    })

    return
  }

  DB.events().forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k:
          v.status = o.value
          break

        case k + '.1':
          v.timestamp = o.value
          break

        case k + '.2':
          v.deviceID = o.value
          break

        case k + '.3':
          v.index = parseInt(o.value)
          break

        case k + '.4':
          v.eventType = o.value
          break

        case k + '.5':
          v.door = o.value
          break

        case k + '.6':
          v.direction = o.value
          break

        case k + '.7':
          v.card = o.value
          break

        case k + '.8':
          v.granted = o.value
          break

        case k + '.9':
          v.reason = o.value
          break

        case k + '.10':
          v.deviceName = o.value
          break

        case k + '.11':
          v.doorName = o.value
          break

        case k + '.12':
          v.cardName = o.value
          break
      }
    }
  })
}

function logs (o) {
  const oid = o.OID

  if (oid === schema.logs.base + '.0.1') {
    DB.tables.logs.first = o.value
    return
  }

  if (oid === schema.logs.base + '.0.2') {
    DB.tables.logs.last = o.value
    return
  }

  if (schema.logs.regex.test(oid)) {
    if (DB.logs().has(oid)) {
      const record = DB.logs().get(oid)
      record.status = o.value
      record.mark = 0
      return
    }

    DB.logs().set(oid, {
      OID: oid,
      timestamp: '',
      uid: '',
      module: {
        type: '',
        ID: '',
        name: '',
        field: ''
      },
      details: '',
      status: o.value,
      mark: 0
    })

    return
  }

  DB.logs().forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k:
          v.status = o.value
          break

        case k + '.1':
          v.timestamp = o.value
          break

        case k + '.2':
          v.uid = o.value
          break

        case k + '.3':
          v.module.type = o.value
          break

        case k + '.4':
          v.module.ID = o.value
          break

        case k + '.5':
          v.module.name = o.value
          break

        case k + '.6':
          v.module.field = o.value
          break

        case k + '.7':
          v.module.details = o.value
          break
      }
    }
  })
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
