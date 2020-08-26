var idleTimer

document.addEventListener('mousedown', event => {
  resetIdle(event)
})

document.addEventListener('click', event => {
  resetIdle(event)
})

document.addEventListener('scroll', event => {
  resetIdle(event)
})

document.addEventListener('keypress', event => {
  resetIdle(event)
})

export async function postAsForm (url = '', data = {}) {
  const pairs = []

  for (const name in data) {
    pairs.push(encodeURIComponent(name) + '=' + encodeURIComponent(data[name]))
  }

  const response = await fetch(url, {
    method: 'POST',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    body: pairs.join('&').replace(/%20/g, '+')
  })

  return response
}

export async function postAsJSON (url = '', data = {}) {
  const response = await fetch(url, {
    method: 'POST',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json' },
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    body: JSON.stringify(data)
  })

  return response
}

export function warning (msg) {
  const message = document.getElementById('message')

  if (message != null) {
    message.innerText = msg
    message.classList.add('warning')
    message.style.display = 'block'
  } else {
    alert(msg)
  }
}

export function onSignOut (event) {
  if (event != null) {
    event.preventDefault()
  }

  postAsJSON('/logout', {})
    .then(response => {
      if (response.status === 200 && response.redirected) {
        window.location = response.url
      } else {
        return response.text()
      }
    })
    .then(msg => {
      warning(msg)
    })
    .catch(function (err) { console.error(err) })
}

export function onIdle () {
  onSignOut()
}

export function resetIdle () {
  if (idleTimer != null) {
    clearTimeout(idleTimer)
  }

  idleTimer = setTimeout(onIdle, 15 * 60 * 1000)
}

export function onCommit (event) {
  const re = /C(.+?)\.commit/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const id = match[1]
    const row = document.getElementById('R' + id)

    if (row && row.hasChildNodes) {
      console.log('commit', row.childNodes)
    }
  }
}

export function onRollback (event) {
  const re = /C(.+?)\.rollback/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const id = match[1]
    const row = document.getElementById('R' + id)
    const commit = document.getElementById('C' + id + '.commit')
    const rollback = document.getElementById('C' + id + '.rollback')

    if (row) {
      const groups = row.querySelectorAll('.group span')
      const re = new RegExp('^G' + id + '.(.+)$')

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
  const re = /G(.+?)\.(.+)/
  const match = event.target.id.match(re)

  if (match.length === 3) {
    const id = match[1]
    const row = document.getElementById('R' + id)
    const commit = document.getElementById('C' + id + '.commit')
    const rollback = document.getElementById('C' + id + '.rollback')
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
