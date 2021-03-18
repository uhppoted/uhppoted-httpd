const db = {
  controllers: new Map()
}

export function UpdateDB (controllers) {
  if (controllers) {
    controllers.forEach(c => {
      update(c)
    })
  }
}

function update (c) {
  const oid = c.OID

  const controller = {
    OID: oid,
    name: c.Name,
    deviceID: c.DeviceID,
    address: c.IP.Address,
    datetime: c.SystemTime.DateTime,
    cards: c.Cards.Records,
    events: c.Events,
    doors: {
      1: c.Doors[1],
      2: c.Doors[2],
      3: c.Doors[3],
      4: c.Doors[4]
    }
  }

  db.controllers.set(oid, controller)
}
