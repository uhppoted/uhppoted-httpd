import { getAsJSON, postAsJSON, warning } from './uhppoted.js'

export function onEdited (event) {
  const input = event.target
  const id = input.dataset.record
  const row = document.getElementById(id)
  const original = input.dataset.original
  const value = input.value

  input.dataset.value = value

  if (value !== original) {
    input.parentElement.dataset.modified = 'true'
  } else {
    delete input.parentElement.dataset.modified
  }

  if (isModified(row)) {
    row.dataset.modified = 'true'
  } else {
    delete row.dataset.modified
  }
}

export function onTick (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)
  const group = document.getElementById(event.target.id)
  const original = group.dataset.original === 'true'
  const value = group.dataset.value === 'true'
  const granted = !value

  group.dataset.value = granted ? 'true' : 'false'
  group.innerText = granted ? 'Y' : 'N'

  if (original !== granted) {
    group.parentElement.dataset.modified = 'true'
  } else {
    delete (group.parentElement.dataset.modified)
  }

  if (isModified(row)) {
    row.dataset.modified = 'true'
  } else {
    delete row.dataset.modified
  }
}

function isModified (row) {
  let modified = false
  Array.from(row.children).forEach((item) => {
    if (item.dataset.modified) {
      modified = true
    }
  })

  return modified
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
      card.dataset.pending = 'true'
    }

    groups.forEach((group) => {
      delete (group.parentElement.dataset.modified)

      if ((group.dataset.record === id) && (group.dataset.value !== group.dataset.original)) {
        update[group.id] = group.dataset.value === 'true'
        group.dataset.pending = 'true'
      }
    })

    delete (row.dataset.modifie)

    const rollback = function () {
      Object.entries(update).forEach(([k, v]) => {
        const e = document.getElementById(k)

        if (e) {
          switch (e.tagName.toLowerCase()) {
            case 'input':
              e.dataset.value = e.dataset.original
              e.value = e.dataset.original
              e.parentElement.dataset.modified = 'true'
              break

            case 'span':
              e.dataset.value = e.dataset.original
              e.innerText = e.dataset.value === 'true' ? 'Y' : 'N'
              e.parentElement.dataset.modified = 'true'
              break
          }
        }
      })
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
            rollback()
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

      delete (card.parentElement.dataset.modified)
    }

    groups.forEach((group) => {
      if (group.dataset.record === id) {
        group.dataset.value = group.dataset.original
        group.innerText = group.dataset.value === 'true' ? 'Y' : 'N'

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
        delete (item.parentElement.dataset.state)
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

  for (const id of rows) {
    const row = document.getElementById(id)

    if (isModified(row)) {
      row.dataset.modified = 'true'
    } else {
      delete row.dataset.modified
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
