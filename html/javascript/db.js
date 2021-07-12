export const DB = {
  interfaces: new Map(),
  controllers: new Map(),

  added: function (tag, recordset) {
    if (recordset) {
      switch (tag) {
        case 'controllers':
          recordset.forEach(r => controller(r, statusToString(r.Status)))
          break
      }
    }
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
          recordset.forEach(r => controller(r, statusToString(r.Status)))
          break
      }
    }
  },

  deleted: function (tag, recordset) {
    if (recordset) {
      switch (tag) {
        case 'controllers':
          recordset.forEach(r => controller(r, 'deleted'))
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
  console.log('DB.object', o)

  // ... update interface property?
  DB.interfaces.forEach((v, k) => {
    if (o.OID.startsWith(k)) {
      switch (o.OID) {
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
  DB.controllers.forEach((v, k) => {
    if (o.OID.startsWith(k)) {
      switch (o.OID) {
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
          v.cards = o.value
          break

        case k + '.6':
          v.events = o.value
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

  // ... add controller ?
  DB.controllers.forEach((v, k) => {
    if (o.OID.startsWith(k)) {
      return
    }

    if (/^0\.1\.1\.[0-9]+$/.test(o.OID)) {
      console.log('>>>> new controller', o.OID)
      controller({ OID: o.OID }, 'unknown')
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

function controller (c, status) {
  const oid = c.OID

  const record = {
    OID: oid,
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

    status: status,
    mark: 0
  }

  if (c.Name) {
    record.name = c.Name
  }

  if (c.DeviceID) {
    record.deviceID = c.DeviceID
  }

  if (c.IP && c.IP.Address) {
    record.address.address = c.IP.Address
    record.address.configured = c.IP.Configured
    record.address.status = statusToString(c.IP.Status)
  }

  if (c.SystemTime) {
    record.datetime.datetime = c.SystemTime.DateTime
    record.datetime.expected = c.SystemTime.Expected
    record.datetime.status = c.SystemTime.Status
  }

  if (c.Cards) {
    record.cards.cards = c.Cards.Records
    record.cards.status = statusToString(c.Cards.Status)
  }

  if (c.Events) {
    record.events.events = c.Events
    record.events.status = 'ok'
  }

  if (c.Doors) {
    record.doors[1] = c.Doors[1]
    record.doors[2] = c.Doors[2]
    record.doors[3] = c.Doors[3]
    record.doors[4] = c.Doors[4]
  }

  if (c.Deleted) {
    record.status = 'deleted'
  }

  DB.controllers.set(oid, record)
}

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

function statusToString (status) {
  switch (status) {
    case 1:
      return 'ok'

    case 2:
      return 'uncertain'

    case 3:
      return 'error'

    case 4:
      return 'unconfigured'

    case 5:
      return 'new'
  }

  return 'unknown'
}
