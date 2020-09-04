import { postAsJSON, warning } from './uhppoted.js'

export function onCommit (event) {
  event.preventDefault()

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

    windmill.dataset.count = (queued + 1).toString()

    postAsJSON('/update', update)
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
        if (group.dataset.value === 'true') {
          group.innerText = 'Y'
        } else {
          group.innerText = 'N'
        }

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

  if (granted) {
    group.dataset.value = 'true'
    group.innerText = 'Y'
  } else {
    group.dataset.value = 'false'
    group.innerText = 'N'
  }

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
