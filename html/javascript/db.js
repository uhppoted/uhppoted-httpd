export const DB = {
  controllers: new Map(),

  added: function (tag, recordset) {
    if (recordset) {
      recordset.forEach(r => update(r, statusToString(r.Status)))
    }
  },

  updated: function (tag, recordset) {
    if (recordset) {
      recordset.forEach(r => update(r, statusToString(r.Status)))
    }
  },

  deleted: function (tag, recordset) {
    if (recordset) {
      recordset.forEach(r => update(r, 'deleted'))
    }
  },

  delete: function (tag, oid) {
    if (oid && this.controllers.has(oid)) {
      let record = this.controllers.get(oid)

      record.status = 'deleted'

      this.controllers.set(oid, record)
    }
  }
}

export function UpdateDB (controllers) {
  if (controllers) {
    controllers.forEach(c => {
      update(c)
    })
  }
}

function update (c, status) {
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

    status: status
  }

  if (c.Name) {
    record.name = c.Name
  }

  if (c.DeviceID) {
    record.deviceID = c.DeviceID
  }

  if (c.IP.Address) {
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
