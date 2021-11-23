export const schema = {
  interfaces: {
    base: '0.1',
    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    type: '.0.3',
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
    regex: /^(0\.2\.[1-9][0-9]*).*$/
  },

  doors: {
    base: '0.3',
    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    regex: /^(0\.3\.([1-9][0-9]*)).*$/
  },

  cards: {
    base: '0.4',
    status: '.0.0',
    created: '.0.1',
    deleted: '.0.2',
    modified: '.0.3',
    name: '.1',
    card: '.2',
    from: '.3',
    to: '.4',
    group: '.5.',
    regex: /^(0\.4\.[1-9][0-9]*).*$/,
    groups: /^(0\.4\.[1-9][0-9]*\.5\.[1-9][0-9]*)(\.[1-3])?$/
  },

  groups: {
    base: '0.5',
    regex: /^0\.5\.([1-9][0-9]*)$/,
    doors: /^(0\.5\.[1-9][0-9]*\.2\.[1-9][0-9]*)(\.[1-3])?$/
  },

  events: {
    base: '0.6',
    regex: /^0\.6\.[1-9][0-9]*$/
  },

  logs: {
    base: '0.7',
    regex: /^0\.7\.[1-9][0-9]*$/
  }
}
