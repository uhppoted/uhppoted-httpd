export const DB = {
  interfaces: new Map(),
  controllers: new Map(),

  added: function (objects) {
    objects.forEach(o => add(o))
  },

  updated: function (tag, recordset) {
    if (recordset) {
      switch (tag) {
        case 'objects':
          recordset.forEach(o => object(o))
          break

        case 'interface':
          iface(recordset)
          break

        case 'controllers':
          throw new Error('OOOPS!! DB.controllers is no longer implemented')
          // recordset.forEach(r => controller(r, statusToString(r.Status)))
          // break
      }
    }
  },

  deleted: function (objects) {
    objects.forEach(o => remove(o))
  },

  delete: function (tag, oid) {
    switch (tag) {
      case 'controllers':
        if (oid && this.controllers.has(oid)) {
          const record = this.controllers.get(oid)

          record.mark = 0
          record.status = 'deleted'
          break
        }
    }
  },

  refreshed: function (tag) {
    mark()
    sweep()
  }
}

function object (o) {
  const oid = o.OID

  // ... interfaces
  if (oid === '0.1.1.0') {
    if (!DB.interfaces.has(oid)) {
      DB.interfaces.set(oid, {
        OID: oid,
        type: o.value,
        name: 'LAN',
        bind: '',
        broadcast: '',
        listen: '',

        status: 'ok',
        mark: 0
      })

      return
    }
  }

  DB.interfaces.forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k + '.1':
          v.name = o.value
          break

        case k + '.2':
          v.bind = o.value
          break

        case k + '.3':
          v.broadcast = o.value
          break

        case k + '.4':
          v.listen = o.value
          break
      }
    }
  })

  // ... update controller property?

  if (/^0\.1\.1\.[1-9][0-9]*$/.test(oid)) {
    if (DB.controllers.has(oid)) {
      DB.controllers.get(oid).status = o.value
      return
    }

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

    return
  }

  DB.controllers.forEach((v, k) => {
    if (oid.startsWith(k)) {
      // INTERIM HACK
      if (v.status === 'new') {
        v.status = 'unknown'
      }

      switch (oid) {
        case k:
          console.log('STATUS: ', o)
          v.status = o.value
          break

        case k + '.0.1':
          v.created = o.value
          break

        case k + '.1':
          v.name = o.value
          break

        case k + '.2':
          v.deviceID = o.value
          break

        case k + '.3':
          v.address = {
            address: o.value,
            configured: o.value,
            status: 'unknown'
          }
          break

        case k + '.4':
          v.datetime = {
            datetime: o.value,
            expected: o.value,
            status: 'unknown'
          }
          break

        case k + '.5':
          v.cards.cards = o.value
          break

        case k + '.6':
          v.events.events = o.value
          break

        case k + '.7':
          v.doors[1] = o.value
          break

        case k + '.8':
          v.doors[2] = o.value
          break

        case k + '.9':
          v.doors[3] = o.value
          break

        case k + '.10':
          v.doors[4] = o.value
          break
      }
    }
  })
}

function add (object) {
  const oid = object.OID

  const controller = {
    OID: oid,
    created: '',
    name: '',
    deviceID: '',

    address: {
      address: '',
      configured: '',
      status: 'unknown'
    },

    datetime: {
      datetime: '',
      expected: '',
      status: 'unknown'
    },

    cards: {
      cards: '',
      status: 'unknown'
    },

    events: {
      events: '',
      status: 'unknown'
    },

    doors: {
      1: '',
      2: '',
      3: '',
      4: ''
    },

    status: 'new',
    mark: 0
  }

  DB.controllers.set(oid, controller)
}

function remove (object) {
  const oid = object.OID

  DB.controllers.forEach((v, k) => {
    if (oid === k) {
      v.status = 'deleted'
    }
  })
}

function iface (c) {
  const oid = c.OID

  const record = {
    OID: oid,
    type: 'LAN',
    name: 'LAN',
    bind: '',
    broadcast: '',
    listen: '',

    status: 'ok',
    mark: 0
  }

  if (c.type) {
    record.type = c.type
  }

  if (c.name && c.name !== '') {
    record.name = c.name
  }

  if (c['bind-address']) {
    record.bind = c['bind-address']
  }

  if (c['broadcast-address']) {
    record.broadcast = c['broadcast-address']
  }

  if (c['listen-address']) {
    record.listen = c['listen-address']
  }

  if (c.deleted) {
    record.status = 'deleted'
  }

  DB.interfaces.set(oid, record)
}

// function controller (c, status) {
//   const oid = c.OID

//   const record = {
//     OID: oid,
//     created: '',
//     name: '',
//     deviceID: '',

//     address: {
//       address: '',
//       configured: '',
//       status: 'unknown'
//     },

//     datetime: {
//       datetime: '',
//       expected: '',
//       status: 'unknown'
//     },

//     cards: {
//       cards: '',
//       status: 'unknown'
//     },

//     events: {
//       events: '',
//       status: 'unknown'
//     },

//     doors: {
//       1: '',
//       2: '',
//       3: '',
//       4: ''
//     },

//     status: status,
//     mark: 0
//   }

//   if (c.Created) {
//     record.created = c.Created
//   }

//   if (c.Name) {
//     record.name = c.Name
//   }

//   if (c.DeviceID) {
//     record.deviceID = c.DeviceID
//   }

//   if (c.IP && c.IP.Address) {
//     record.address.address = c.IP.Address
//     record.address.configured = c.IP.Configured
//     record.address.status = statusToString(c.IP.Status)
//   }

//   if (c.SystemTime) {
//     record.datetime.datetime = c.SystemTime.DateTime
//     record.datetime.expected = c.SystemTime.Expected
//     record.datetime.status = c.SystemTime.Status
//   }

//   if (c.Cards) {
//     record.cards.cards = c.Cards.Records
//     record.cards.status = statusToString(c.Cards.Status)
//   }

//   if (c.Events) {
//     record.events.events = c.Events
//     record.events.status = 'ok'
//   }

//   if (c.Doors) {
//     record.doors[1] = c.Doors[1]
//     record.doors[2] = c.Doors[2]
//     record.doors[3] = c.Doors[3]
//     record.doors[4] = c.Doors[4]
//   }

//   DB.controllers.set(oid, record)
// }

function mark () {
  DB.controllers.forEach(v => {
    v.mark += 1
  })
}

function sweep () {
  DB.controllers.forEach((v, k) => {
    if (v.mark >= 25 && v.status === 'deleted') {
      DB.controllers.delete(k)
    }
  })
}

// function statusToString (status) {
//   switch (status) {
//     case 1:
//       return 'ok'

//     case 2:
//       return 'uncertain'

//     case 3:
//       return 'error'

//     case 4:
//       return 'unconfigured'

//     case 5:
//       return 'new'
//   }

//   return 'unknown'
// }
