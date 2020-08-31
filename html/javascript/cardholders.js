import { postAsJSON, warning } from './uhppoted.js'

export function onCommit (event) {
  event.preventDefault()

  const id = event.target.dataset.record
  const row = document.getElementById(id)
  const commit = document.getElementById(id + '_commit')
  const rollback = document.getElementById(id + '_rollback')

  if (row) {
    const update = {}
    const groups = row.querySelectorAll('.group span')

    groups.forEach((group) => {
      if ((group.dataset.record === id) && (group.dataset.value !== group.dataset.original)) {
        update[group.id] = group.dataset.value === 'true'
      }
    })

    postAsJSON('/update', update)
      .then(response => {
        switch (response.status) {
          case 200:
            response.json().then(object => {
              updated(object)

              delete (row.dataset.modified)
              commit.dataset.enabled = 'false'
              rollback.dataset.enabled = 'false'
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
  const commit = document.getElementById(id + '_commit')
  const rollback = document.getElementById(id + '_rollback')

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

        delete (group.dataset.modified)
      }
    })

    delete (row.dataset.modified)
    commit.dataset.enabled = 'false'
    rollback.dataset.enabled = 'false'
  }
}

export function onTick (event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)
  const commit = document.getElementById(id + '_commit')
  const rollback = document.getElementById(id + '_rollback')
  const group = document.getElementById(event.target.id)
  const original = group.dataset.original === 'true'
  const value = group.dataset.value === 'true'
  const granted = !value
  let modified = (row.dataset.modified && parseInt(row.dataset.modified)) | 0

  if (granted) {
    group.dataset.value = 'true'
    group.innerText = 'Y'
  } else {
    group.dataset.value = 'false'
    group.innerText = 'N'
  }

  if (original !== granted) {
    group.dataset.modified = 'true'
    modified += 1
  } else {
    delete (group.dataset.modified)
    modified -= 1
  }

  if (modified > 0) {
    row.dataset.modified = modified.toString()
    commit.dataset.enabled = 'true'
    rollback.dataset.enabled = 'true'
  } else {
    delete row.dataset.modified
    commit.dataset.enabled = 'false'
    rollback.dataset.enabled = 'false'
  }
}

function updated (list) {
  for (const [k, v] of Object.entries(list)) {
    const item = document.getElementById(k)

    if (item) {
      item.dataset.original = v
      item.dataset.value = v
      item.innerHTML = v ? 'Y' : 'N'

      delete (item.dataset.modified)
    }
  }
}
