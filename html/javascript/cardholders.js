import { getAsJSON, postAsJSON, warning, dismiss } from './uhppoted.js'

export function onEdited (event) {
  set(event.target, event.target.value)
}

export function onTick (event) {
  const value = !(event.target.dataset.value === 'true')
  event.target.innerText = value ? 'Y' : 'N'

  set(event.target, value)
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
    const card = row.querySelector('.cardnumber input')
    const groups = row.querySelectorAll('.group span')
    const list = [card, ...groups]

    list.forEach((item) => {
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

    const windmill = document.getElementById('windmill')
    const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

    windmill.dataset.count = (queued + 1).toString()

    postAsJSON('/cardholders', update)
      .then(response => {
        const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)
        if (queued > 1) {
          windmill.dataset.count = (queued - 1).toString()
        } else {
          delete (windmill.dataset.count)
        }

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
        console.log(err)
      })
  }
}

export function onRollback (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  if (row) {
    const card = row.querySelector('.cardnumber input')
    const groups = row.querySelectorAll('.group span')
    const list = [card, ...groups]

    list.forEach((item) => {
      if ((item.dataset.record === id) && (item.dataset.value !== item.dataset.original)) {
        item.dataset.value = item.dataset.original
        item.parentElement.dataset.state = ''

        switch (item.tagName.toLowerCase()) {
          case 'input':
            item.value = item.dataset.value
            break

          case 'span':
            item.innerHTML = item.dataset.value ? 'Y' : 'N'
            break
        }
      }
    })

    delete (row.dataset.modified)
  }
}

export function onRefresh (event) {
  const windmill = document.getElementById('windmill')
  const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

  windmill.dataset.count = (queued + 1).toString()

  dismiss()

  getAsJSON('/cardholders')
    .then(response => {
      const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)
      if (queued > 1) {
        windmill.dataset.count = (queued - 1).toString()
      } else {
        delete (windmill.dataset.count)
      }

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
        switch (item.tagName.toLowerCase()) {
          case 'input':
            item.value = v
            break

          case 'span':
            item.innerHTML = v ? 'Y' : 'N'
            break
        }
      }

      if (item.dataset.value !== v.toString()) {
        item.parentElement.dataset.state = 'conflict'
      }

      if (item.parentElement.dataset === 'pending') {
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

    const card = document.getElementById(record.Card.ID)
    if (card) {
      card.value = record.Card.Number
      set(card, record.Card.Number)
    }

    record.Groups.forEach((group) => {
      const g = document.getElementById(group.ID)

      if (g) {
        g.innerHTML = group.Value ? 'Y' : 'N'
        set(g, group.Value)
      }
    })
  })
}
