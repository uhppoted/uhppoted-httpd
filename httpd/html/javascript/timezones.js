// Ref. https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

const LOCALE = 'en-ca-iso8601'

const DEFAULT = new Intl.DateTimeFormat(LOCALE, {
  timeZone: undefined,
  timeZoneName: 'short',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: 'numeric',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
})

export const timezones = new Map([
  ['local', local],
  ['UTC', short],
  ['GMT', short],
  ['PST', short],
  ['PST8PDT', short],
  ['Africa/Cairo', long],
  ['Etc/GMT-2', short],
])

function local(dt, _tz) {
  try {
    const fmt = new Intl.DateTimeFormat(LOCALE, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: 'numeric',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })

    return fmt.format(dt).replaceAll(',', '')
  } catch {
    return DEFAULT.format(dt).replaceAll(',', '')
  }
}

function short(dt, tz) {
  try {
    const fmt = new Intl.DateTimeFormat(LOCALE, {
      timeZone: tz,
      timeZoneName: 'short',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: 'numeric',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })

    return fmt.format(dt).replaceAll(',', '')
  } catch {
    return DEFAULT.format(dt).replaceAll(',', '')
  }
}

function long(dt, tz) {
  try {
    const fmt = new Intl.DateTimeFormat(LOCALE, {
      timeZone: tz,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: 'numeric',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })

    const formatted = fmt.format(dt).replaceAll(',', '')

    return `${formatted} ${tz}`
  } catch {
    return DEFAULT.format(dt).replaceAll(',', '')
  }
}
