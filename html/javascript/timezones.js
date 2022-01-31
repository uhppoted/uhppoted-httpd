// Ref. https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

export const timezones = new Map([
  ['UTC', format],
  ['GMT', format],
  ['PST', format],
  ['PST8PDT', format],
  ['Africa/Cairo', format],
  ['Etc/GMT-2', format]
])

function format (dt, tz) {
  const fmt = new Intl.DateTimeFormat('en-ca-iso8601', {
    timeZone: tz,
    timeZoneName: 'short',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: 'numeric',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })

  return fmt.format(dt).replaceAll(',', '')
}
