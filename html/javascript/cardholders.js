import { getAsJSON, postAsJSON, warning } from './uhppoted.js'

export function onCommit (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  if (row) {
    const update = {}
    const groups = row.querySelectorAll('.group span')
    const windmill = document.getElementById('windmill')
    const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

    groups.forEach((group) => {
      delete (group.dataset.edited)

      if ((group.dataset.record === id) && (group.dataset.value !== group.dataset.original)) {
        update[group.id] = group.dataset.value === 'true'
        group.dataset.pending = 'true'
      }
    })

    delete (row.dataset.edited)

    const rollback = function () {
      Object.entries(update).forEach(([k, v]) => {
        const g = document.getElementById(k)

        if (g) {
          g.dataset.value = g.dataset.original
          g.dataset.modified = 'true'
          g.innerText = g.dataset.value === 'true' ? 'Y' : 'N'
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
    const groups = row.querySelectorAll('.group span')

    groups.forEach((group) => {
      if (group.dataset.record === id) {
        group.dataset.value = group.dataset.original
        group.innerText = group.dataset.value === 'true' ? 'Y' : 'N'

        delete (group.dataset.edited)
      }
    })

    delete (row.dataset.edited)
  }
}

export function onTick (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)
  const group = document.getElementById(event.target.id)
  const original = group.dataset.original === 'true'
  const value = group.dataset.value === 'true'
  const granted = !value
  let edited = (row.dataset.edited && parseInt(row.dataset.edited)) | 0

  group.dataset.value = granted ? 'true' : false
  group.innerText = granted ? 'Y' : 'N'

  if (original !== granted) {
    group.dataset.edited = 'true'
    edited += 1
  } else {
    delete (group.dataset.edited)
    edited -= 1
  }

  if (edited > 0) {
    row.dataset.edited = edited.toString()
  } else {
    delete row.dataset.edited
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
  for (const [k, v] of Object.entries(list)) {
    const item = document.getElementById(k)

    if (item) {
      if (item.dataset.value !== v.toString()) {
        item.dataset.modified = 'true'
      } else {
        delete (item.dataset.modified)
      }

      item.dataset.original = v
      item.dataset.value = v
      item.innerHTML = v ? 'Y' : 'N'

      delete (item.dataset.edited)
    }
  }
}

function refresh (db) {
  const records = db.cardholders

  records.forEach((record) => {
    const row = document.getElementById(record.ID)

    if (row) {
      delete (row.dataset.edited)
    }

    record.Groups.forEach((group) => {
      const v = group.Value
      const g = document.getElementById(group.ID)

      if (g) {
        g.dataset.original = v
        g.dataset.value = v
        g.innerHTML = v ? 'Y' : 'N'

        delete (g.dataset.modified)
        delete (g.dataset.edited)
      }
    })
  })
}
