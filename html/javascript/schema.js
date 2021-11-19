export const schema = {
  interfaces: {
    base: '0.1',
    status: '.0.0',
    type: '.1',
    name: '.2',
    bind: '.3',
    broadcast: '.4',
    listen: '.5',

    regex: /^(0\.1\.[1-9][0-9]*).*$/
  },

  controllers: {
    base: '0.2',
    regex: /^0\.2\.[1-9][0-9]*$/
  },

  doors: {
    base: '0.3',
    regex: /^0\.3\.([1-9][0-9]*)$/
  },

  cards: {
    base: '0.4',
    regex: /^0\.4\.[1-9][0-9]*$/,
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
