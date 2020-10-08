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

  element.dataset.value = v

  if (v !== original) {
    element.parentElement.classList.add('modified')
  } else {
    element.parentElement.classList.remove('modified')
  }

  const unmodified = Array.from(row.children).every(item => !item.classList.contains('modified'))
  if (unmodified) {
    row.classList.remove('modified')
  } else {
    row.classList.add('modified')
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

        item.parentElement.classList.add('pending')
        item.parentElement.classList.remove('modified')
      }
    })

    row.classList.remove('modified')

    const reset = function () {
      Object.entries(update).forEach(([k, v]) => {
        document.getElementById(k).parentElement.classList.remove('pending')
        document.getElementById(k).parentElement.classList.add('modified')
      })

      row.classList.add('modified')
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
        item.parentElement.classList.remove('modified')

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

    row.classList.remove('modified')
  }
}

export function onAdd (event) {
  const tbody = document.getElementById('table').querySelector('table tbody')

  if (tbody) {
    const row = tbody.insertRow()
    const name = row.insertCell()
    const controls = row.insertCell()
    const card = row.insertCell()
    const from = row.insertCell()
    const to = row.insertCell()
    const groups = []

    // eslint-disable-next-line
    {{ range.db.Groups }} 
    // eslint-disable-next-line
    groups.push(row.insertCell())
    // eslint-disable-next-line
    {{ end }}

    name.classList.add('name')
    controls.classList.add('controls')
    card.classList.add('card')
    from.classList.add('from')
    to.classList.add('to')
    groups.forEach(g => g.classList.add('group'))

    name.innerHTML = '<img class="flag" src="images/corner.svg" />' +
                     '<input class="field" type="text" value="" onchange="onEdited(event)" data-record="" data-original="" data-value="" />'
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
    update(document.getElementById(k), v)
  }
}

function refresh (db) {
  const records = db.cardholders

  records.forEach((record) => {
    const name = document.getElementById(record.Name.ID)
    const card = document.getElementById(record.Card.ID)
    const from = document.getElementById(record.From.ID)
    const to = document.getElementById(record.To.ID)

    update(name, record.Name.Name)
    update(card, record.Card.Number)
    update(from, record.From.Date)
    update(to, record.To.Date)

    record.Groups.forEach((group) => {
      update(document.getElementById(group.ID), group.Value)
    })
  })
}

function update (element, value) {
  const v = value.toString()

  if (element) {
    element.dataset.original = v

    // check for conflicts with concurrently modified fields

    if (element.parentElement.classList.contains('modified')) {
      element.parentElement.classList.remove('pending')

      if (element.dataset.value !== v.toString()) {
        element.parentElement.classList.add('conflict')
      } else {
        element.parentElement.classList.remove('modified')
        element.parentElement.classList.remove('conflict')
      }

      return
    }

    // mark fields with unexpected values after submit

    if (element.parentElement.classList.contains('pending')) {
      element.parentElement.classList.remove('pending')

      if (element.dataset.value !== v.toString()) {
        element.parentElement.classList.add('conflict')
      } else {
        element.parentElement.classList.remove('conflict')
      }
    }

    // update unmodified fields

    switch (element.getAttribute('type').toLowerCase()) {
      case 'text':
      case 'number':
      case 'date':
        element.value = v
        break

      case 'checkbox':
        element.checked = (v === 'true')
        break
    }

    set(element, value)
  }
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
