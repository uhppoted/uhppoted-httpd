import { getAsJSON, postAsJSON, warning, dismiss } from './uhppoted.js'

export function onEdited (event) {
  set(event.target, event.target.value)
}

export function onTick (event) {
  set(event.target, event.target.checked)
}

function set (element, value) {
  const rowid = element.dataset.record
  const row = document.getElementById(rowid)
  const original = element.dataset.original
  const v = value.toString()
  let modified = false

  element.dataset.value = v

  if (v !== original) {
    element.parentElement.dataset.state = 'modified'
  } else {
    element.parentElement.dataset.state = ''
  }

  Array.from(row.children).forEach((item) => {
    if (item.dataset.state === 'modified') {
      modified = true
    }
  })

  if (modified) {
    row.dataset.modified = 'true'
  } else {
    delete row.dataset.modified
  }
}

export function onCommit (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  if (row) {
    const update = {}
    const fields = row.querySelectorAll('.field')

    fields.forEach((item) => {
      if ((item.dataset.record === id) && (item.dataset.value !== item.dataset.original)) {
        update[item.id] = item.dataset.value

        item.parentElement.dataset.state = 'pending'
      }
    })

    delete (row.dataset.modified)

    const reset = function () {
      Object.entries(update).forEach(([k, v]) => {
        document.getElementById(k).parentElement.dataset.state = 'modified'
      })

      row.dataset.modified = 'true'
    }

    busy()

    postAsJSON('/cardholders', update)
      .then(response => {
        unbusy()

        switch (response.status) {
          case 200:
            response.json().then(object => { updated(object.db.updated) })
            break

          default:
            reset()
            response.text().then(message => { warning(message) })
        }
      })
      .catch(function (err) {
        unbusy()
        reset()
        warning(`Error committing update (ERR:${err.message.toLowerCase()})`)
      })
  }
}

export function onRollback (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  if (row) {
    const fields = row.querySelectorAll('.field')

    fields.forEach((item) => {
      if ((item.dataset.record === id) && (item.dataset.value !== item.dataset.original)) {
        item.dataset.value = item.dataset.original
        item.parentElement.dataset.state = ''

        switch (item.getAttribute('type').toLowerCase()) {
          case 'text':
          case 'number':
          case 'date':
            item.value = item.dataset.value
            break

          case 'checkbox':
            item.checked = item.dataset.value === 'true'
            break
        }
      }
    })

    delete (row.dataset.modified)
  }
}

export function onRefresh (event) {
  busy()
  dismiss()

  getAsJSON('/cardholders')
    .then(response => {
      unbusy()

      switch (response.status) {
        case 200:
          response.json().then(object => { refresh(object.db) })
          break

        default:
          response.text().then(message => { warning(message) })
      }
    })
    .catch(function (err) {
      console.log(err)
    })
}

function updated (list) {
  for (const [k, v] of Object.entries(list)) {
    const item = document.getElementById(k)

    if (item) {
      item.dataset.original = v

      if (item.parentElement.dataset.state !== 'modified') {
        switch (item.getAttribute('type').toLowerCase()) {
          case 'text':
          case 'number':
          case 'date':
            item.value = v
            break

          case 'checkbox':
            item.checked = item.dataset.value === 'true'
            break
        }
      }

      if (item.dataset.value !== v.toString()) {
        item.parentElement.dataset.state = 'conflict'
      }

      if (item.parentElement.dataset.state === 'pending') {
        item.parentElement.dataset.state = ''
      }
    }
  }
}

function refresh (db) {
  const records = db.cardholders

  records.forEach((record) => {
    const row = document.getElementById(record.ID)

    if (row) {
      delete (row.dataset.modified)
    }

    const name = document.getElementById(record.Name.ID)
    if (name) {
      name.value = record.Name.Name
      set(name, record.Name.Name)
    }

    const card = document.getElementById(record.Card.ID)
    if (card) {
      card.value = record.Card.Number
      set(card, record.Card.Number)
    }

    const from = document.getElementById(record.From.ID)
    if (from) {
      from.value = record.From.Date
      set(from, record.From.Date)
    }

    const to = document.getElementById(record.To.ID)
    if (to) {
      to.value = record.To.Date
      set(to, record.To.Date)
    }

    record.Groups.forEach((group) => {
      const g = document.getElementById(group.ID)

      if (g) {
        g.checked = group.Value
        set(g, group.Value)
      }
    })
  })
}

function busy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  windmill.dataset.count = (queued + 1).toString()
}

function unbusy () {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  if (queued > 1) {
    windmill.dataset.count = (queued - 1).toString()
  } else {
    delete (windmill.dataset.count)
  }
}
