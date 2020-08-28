import { postAsJSON, warning } from './uhppoted.js'

export function onCommit (event) {
  event.preventDefault()

  const re = /C(.+?)_commit/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const id = match[1]
    const row = document.getElementById('R' + id)
    const commit = document.getElementById('C' + id + '_commit')
    const rollback = document.getElementById('C' + id + '_rollback')

    if (row) {
      const groups = row.querySelectorAll('.group span')
      const re = new RegExp('^G' + id + '_(.+)$')
      const update = {}

      groups.forEach((group) => {
        const match = group.id.match(re)

        if ((match.length === 2) && (group.dataset.value !== group.dataset.original)) {
          update[group.id] = group.dataset.value === 'true'
        }
      })

      postAsJSON('/update', update)
        .then(response => {
          if (response.status === 200) {
            delete (row.dataset.modified)
            commit.dataset.enabled = 'false'
            rollback.dataset.enabled = 'false'

            Object.entries(update).forEach(([k, v]) => {
              const item = row.querySelector('#' + k)

              if (item) {
                // item.dataset.original = v
                // item.dataset.value = v
                item.innerHTML = '?'
                delete (item.dataset.modified)
              }
            })

            return ''
          } else {
            return response.text()
          }
        })
        .then(msg => {
          if (msg !== '') {
            warning(msg)
          }
        })
        .catch(function (err) {
          console.log(err)
        })
    }
  }
}

export function onRollback (event) {
  const re = /C(.+?)_rollback/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const id = match[1]
    const row = document.getElementById('R' + id)
    const commit = document.getElementById('C' + id + '_commit')
    const rollback = document.getElementById('C' + id + '_rollback')

    if (row) {
      const groups = row.querySelectorAll('.group span')
      const re = new RegExp('^G' + id + '_(.+)$')

      groups.forEach((group) => {
        const match = group.id.match(re)

        if (match.length === 2) {
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
}

export function onTick (event) {
  const re = /G(.+?)_(.+)/
  const match = event.target.id.match(re)

  if (match.length === 3) {
    const id = match[1]
    const row = document.getElementById('R' + id)
    const commit = document.getElementById('C' + id + '_commit')
    const rollback = document.getElementById('C' + id + '_rollback')
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
}
