import { update, trim } from './tabular.js'
import { DB, alive } from './db.js'
import { schema } from './schema.js'

const pagesize = 5

export function refreshed () {
  const users = [...DB.users().values()]
    .filter(u => alive(u))
    .sort((p, q) => p.created.localeCompare(q.created))

  realize(users)

  // renders a 'page size' chunk of users
  const f = function (offset) {
    let ix = offset
    let count = 0
    while (count < pagesize && ix < users.length) {
      const o = users[ix]
      const row = updateFromDB(o.OID, o)
      if (row) {
        if (o.status === 'new') {
          row.classList.add('new')
        } else {
          row.classList.remove('new')
        }
      }

      count++
      ix++
    }
  }

  // sorts the table rows by 'created'
  const g = function () {
    const focused = document.activeElement

    if (!focused || focused.nodeName !== 'INPUT') {
      const table = document.querySelector('#users table')
      const tbody = table.tBodies[0]

      tbody.sort((p, q) => {
        const u = DB.users().get(p.dataset.oid)
        const v = DB.users().get(q.dataset.oid)

        return u.created.localeCompare(v.created)
      })
    }
  }

  const chunk = offset => new Promise(resolve => {
    f(offset)
    resolve(true)
  })

  async function * render () {
    for (let ix = 0; ix < users.length; ix += pagesize) {
      yield chunk(ix).then(() => ix)
    }
  }

  (async function loop () {
    for await (const _ of render()) {
      // empty
    }
  })()
    .then(() => g())
    .catch(err => console.error(err))
}

export function deletable (row) {
  const name = row.querySelector('td input.name')
  const uid = row.querySelector('td input.uid')
  const re = /^\s*$/

  if (name && name.dataset.oid !== '' && re.test(name.dataset.value) &&
      uid && uid.dataset.oid !== '' && re.test(uid.dataset.value)) {
    return true
  }

  return false
}

function realize (users) {
  const table = document.querySelector('#users table')
  const tbody = table.tBodies[0]

  trim('users', users, tbody.querySelectorAll('tr.user'))

  users.forEach(o => {
    let row = tbody.querySelector("tr[data-oid='" + o.OID + "']")
    if (!row) {
      row = add(o.OID, o)
    }
  })
}

function add (oid, record) {
  const uuid = 'R' + oid.replaceAll(/[^0-9]/g, '')
  const tbody = document.getElementById('users').querySelector('table tbody')

  if (tbody) {
    const template = document.querySelector('template#user')
    const row = tbody.insertRow()

    row.id = uuid
    row.classList.add('user')
    row.dataset.oid = oid
    row.dataset.status = 'unknown'
    row.innerHTML = template.innerHTML

    const commit = row.querySelector('td span.commit')
    commit.id = uuid + '_commit'
    commit.dataset.record = uuid

    const rollback = row.querySelector('td span.rollback')
    rollback.id = uuid + '_rollback'
    rollback.dataset.record = uuid

    const fields = [
      { suffix: 'name', oid: `${oid}${schema.users.name}`, selector: 'td input.name' },
      { suffix: 'uid', oid: `${oid}${schema.users.uid}`, selector: 'td input.uid' },
      { suffix: 'role', oid: `${oid}${schema.users.role}`, selector: 'td input.role' },
      { suffix: 'password', oid: `${oid}${schema.users.password}`, selector: 'td input.password' },
      { suffix: 'otp', oid: `${oid}${schema.users.otp}`, selector: 'td label.otp input' },
      { suffix: 'locked', oid: `${oid}${schema.users.locked}`, selector: 'td label.locked input' }
    ]

    fields.forEach(f => {
      const field = row.querySelector(f.selector)
      if (field) {
        field.id = uuid + '-' + f.suffix
        field.value = ''
        field.dataset.oid = f.oid
        field.dataset.record = uuid
        field.dataset.original = ''
        field.dataset.value = ''
      } else {
        console.error(f)
      }
    })

    return row
  }
}

function updateFromDB (oid, record) {
  const row = document.querySelector("div#users tr[data-oid='" + oid + "']")

  const name = row.querySelector(`[data-oid="${oid}${schema.users.name}"]`)
  const uid = row.querySelector(`[data-oid="${oid}${schema.users.uid}"]`)
  const role = row.querySelector(`[data-oid="${oid}${schema.users.role}"]`)
  const password = row.querySelector(`[data-oid="${oid}${schema.users.password}"]`)
  const otp = row.querySelector(`[data-oid="${oid}${schema.users.otp}"]`)
  const locked = row.querySelector(`[data-oid="${oid}${schema.users.locked}"]`)

  row.dataset.status = record.status

  update(name, record.name)
  update(uid, record.uid)
  update(role, record.role)
  update(password, record.password)
  update(otp, record.otp)
  update(locked, record.locked)

  if (record.otp === 'true') {
    otp.disabled = false
    otp.parentElement.classList.add('visible')
  } else {
    otp.disabled = true
    otp.parentElement.classList.remove('visible')
  }

  if (record.locked === 'true') {
    locked.disabled = false
    locked.parentElement.classList.add('visible')
  } else {
    locked.disabled = true
    locked.parentElement.classList.remove('visible')
  }

  return row
}
