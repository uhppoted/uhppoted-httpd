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

  element.dataset.value = v

  if (v !== original) {
    element.parentElement.dataset.state = 'modified'
  } else {
    element.parentElement.dataset.state = ''
  }

  let modified = false
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
    const windmill = document.getElementById('windmill')
    const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

    delete (card.parentElement.dataset.modified)

    if ((card.dataset.record === id) && (card.dataset.value !== card.dataset.original)) {
      update[card.id] = card.value
      card.dataset.value = card.value
      card.dataset.pending = 'true'
      card.parentElement.dataset.state = 'pending'
    }

    groups.forEach((group) => {
      delete (group.parentElement.dataset.modified)

      if ((group.dataset.record === id) && (group.dataset.value !== group.dataset.original)) {
        update[group.id] = group.dataset.value === 'true'
        group.dataset.pending = 'true'
        group.parentElement.dataset.state = 'pending'
      }
    })

    delete (row.dataset.modified)

    const reset = function () {
      Object.entries(update).forEach(([k, v]) => {
        const e = document.getElementById(k)

        if (e) {
          switch (e.tagName.toLowerCase()) {
            case 'input':
              e.parentElement.dataset.state = 'modified'
              e.parentElement.dataset.modified = 'true'
              break

            case 'span':
              e.parentElement.dataset.state = 'modified'
              e.parentElement.dataset.modified = 'true'
              break
          }
        }
      })

      row.dataset.modified = 'true'
    }

    windmill.dataset.count = (queued + 1).toString()

    postAsJSON('/cardholders', update)
      .then(response => {
        const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)
        if (queued > 1) {
          windmill.dataset.count = (queued - 1).toString()
        } else {
          delete (windmill.dataset.count)
        }

        groups.forEach((group) => {
          delete (group.dataset.pending)
        })

        switch (response.status) {
          case 200:
            response.json().then(object => {
              updated(object.db.updated)
            })
            break

          default:
            reset()
            response.text().then(message => {
              warning(message)
            })
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

    if (card.dataset.record === id) {
      card.dataset.value = card.dataset.original
      card.value = card.dataset.original

      card.parentElement.dataset.state = ''
      delete (card.parentElement.dataset.modified)
    }

    groups.forEach((group) => {
      if (group.dataset.record === id) {
        group.dataset.value = group.dataset.original
        group.innerText = group.dataset.value === 'true' ? 'Y' : 'N'

        group.parentElement.dataset.state = ''
        delete (group.parentElement.dataset.modified)
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
          response.json().then(object => {
            refresh(object.db)
          })
          break

        default:
          response.text().then(message => {
            warning(message)
          })
      }
    })
    .catch(function (err) {
      console.log(err)
    })
}

function updated (list) {
  const rows = new Set()

  for (const [k, v] of Object.entries(list)) {
    const item = document.getElementById(k)

    if (item) {
      if (item.dataset.value !== v.toString()) {
        item.parentElement.dataset.state = 'conflict'
      } else {
        item.parentElement.dataset.state = ''
      }

      item.dataset.original = v
      item.dataset.value = v

      switch (item.tagName.toLowerCase()) {
        case 'input':
          item.value = v
          break

        case 'span':
          item.innerHTML = v ? 'Y' : 'N'
          break
      }

      rows.add(item.dataset.record)
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
      const v = record.Card.Number

      card.dataset.original = v
      card.dataset.value = v
      card.value = v

      delete (card.parentElement.dataset.modified)
      delete (card.parentElement.dataset.state)
    }

    record.Groups.forEach((group) => {
      const v = group.Value
      const g = document.getElementById(group.ID)

      if (g) {
        g.dataset.original = v
        g.dataset.value = v
        g.innerHTML = v ? 'Y' : 'N'

        delete (g.parentElement.dataset.modified)
        delete (g.parentElement.dataset.state)
      }
    })
  })
}
