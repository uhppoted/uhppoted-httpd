export const schema = {
  system: {
    base: '0.0',
    cards: {
      base: '.1',
      defaultStartDate: '.1.1',
      defaultEndDate: '.1.2'
    }
  },

  interfaces: {
    base: '0.1',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.4',

    name: '.1',
    bind: '.3.1',
    broadcast: '.3.2',
    listen: '.3.3',

    regex: /^(0\.1\.[1-9][0-9]*).*$/
  },

  controllers: {
    base: '0.2',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.4',

    name: '.1',
    deviceID: '.2',
    endpoint: {
      status: '.3.0',
      address: '.3.1',
      configured: '.3.2',
      protocol: '.3.3'
    },
    datetime: {
      status: '.4.0',
      current: '.4.1',
      configured: '.4.2'
    },
    cards: {
      status: '.5.0',
      count: '.5.1'
    },
    events: {
      status: '.6.0',
      first: '.6.1',
      last: '.6.2',
      current: '.6.3'
    },
    door1: '.7.1',
    door2: '.7.2',
    door3: '.7.3',
    door4: '.7.4',
    interlock: '.8',

    regex: /^(0\.2\.[1-9][0-9]*).*$/
  },

  doors: {
    base: '0.3',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.4',

    regex: /^(0\.3\.([1-9][0-9]*)).*$/
  },

  cards: {
    base: '0.4',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.4',

    name: '.1',
    card: '.2',
    from: '.3',
    to: '.4',
    group: '.5.',
    // {{if .WithPIN}}
    PIN: '.6',
    // {{end}}

    regex: /^(0\.4\.[1-9][0-9]*).*$/,
    groups: /^(0\.4\.[1-9][0-9]*\.5\.[1-9][0-9]*)(\.[1-3])?$/
  },

  groups: {
    base: '0.5',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.4',

    name: '.1',
    door: '.2',

    regex: /^(0\.5\.([1-9][0-9]*)).*$/,
    doors: /^(0\.5\.[1-9][0-9]*\.2\.[1-9][0-9]*)(\.[1-3])?$/
  },

  events: {
    base: '0.6',

    status: '.0.0',
    first: '.0.1',
    last: '.0.2',

    timestamp: '.1',
    deviceID: '.2',
    index: '.3',
    type: '.4',
    door: '.5',
    direction: '.6',
    card: '.7',
    granted: '.8',
    reason: '.9',
    deviceName: '.10',
    doorName: '.11',
    cardName: '.12',

    regex: /^(0\.6\.[1-9][0-9]*).*$/
  },

  logs: {
    base: '0.7',

    first: '.0.1',
    last: '.0.2',

    timestamp: '.1',
    uid: '.2',
    item: '.3',
    itemID: '.4',
    itemName: '.5',
    field: '.6',
    details: '.7',

    regex: /^(0\.7\.[1-9][0-9]*).*$/
  },

  users: {
    base: '0.8',

    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',

    name: '.1',
    uid: '.2',
    role: '.3',
    password: '.4',
    otp: '.5',
    locked: '.6',

    regex: /^(0\.8\.[1-9][0-9]*).*$/
  }
}
