export const DB = {
  interfaces: new Map(),
  controllers: new Map(),
  doors: new Map(),

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
  if (oid === '0.1.1.1.1') {
    if (!DB.interfaces.has(oid)) {
      DB.interfaces.set(oid, {
        OID: oid,
        type: 'LAN',
        name: 'LAN',
        bind: '',
        broadcast: '',
        listen: '',

        status: o.value,
        mark: 0
      })

      return
    }
  }

  DB.interfaces.forEach((v, k) => {
    if (oid.startsWith(k)) {
      switch (oid) {
        case k + '.0':
          v.type = o.value
          break

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

  // ... controllers

  if (/^0\.1\.1\.2\.[1-9][0-9]*$/.test(oid)) {
    if (DB.controllers.has(oid)) {
      const record = DB.controllers.get(oid)
      record.status = o.value
      record.mark = 0
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
          v.address.address = o.value
          break

        case k + '.3.1':
          v.address.configured = o.value
          break

        case k + '.3.2':
          v.address.status = o.value
          break

        case k + '.4':
          v.datetime.datetime = o.value
          break

        case k + '.4.1':
          v.datetime.expected = o.value
          break

        case k + '.4.2':
          v.datetime.status = o.value
          break

        case k + '.5':
          v.cards.cards = o.value
          break

        case k + '.5.1':
          v.cards.status = o.value
          break

        case k + '.6':
          v.events.events = o.value
          break

        case k + '.6.1':
          v.events.status = o.value
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

  // ... doors

  if (/^0\.3.[1-9][0-9]*$/.test(oid)) {
    if (DB.doors.has(oid)) {
      const record = DB.doors.get(oid)
      record.status = o.value
      record.mark = 0
      return
    }

    DB.doors.set(oid, {
      OID: oid,
      created: '',
      name: '',
      status: o.value,
      mark: 0
    })

    return
  }

  DB.doors.forEach((v, k) => {
    if (oid.startsWith(k)) {
      // INTERIM HACK
      if (v.status === 'new') {
        v.status = 'unknown'
      }

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
      }
    }
  })
}

function mark () {
  DB.controllers.forEach(v => {
    v.mark += 1
  })

  DB.doors.forEach(v => {
    v.mark += 1
  })
}

function sweep () {
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
}
