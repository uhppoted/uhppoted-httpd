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
  const re = /C(.*?)\.commit/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const cid = match[1]
    const row = document.getElementById('R' + cid)

    if (row && row.hasChildNodes) {
      console.log('commit', row.childNodes)
    }
  }
}

export function onRollback (event) {
  const re = /C(.*?)\.rollback/
  const match = event.target.id.match(re)

  if (match.length === 2) {
    const cid = match[1]
    const row = document.getElementById('R' + cid)

    if (row && row.hasChildNodes) {
      console.log('rollback', row.childNodes)
    }
  }
}

export function onTick (event) {
  const re = /G(.*?)\.(.*)/
  const match = event.target.id.match(re)

  if (match.length === 3) {
    const cid = match[1]
    const commit = document.getElementById('C' + cid + '.commit')
    const rollback = document.getElementById('C' + cid + '.rollback')
    const group = document.getElementById(event.target.id)
    const original = group.dataset.original === 'true'
    const value = group.dataset.value === 'true'
    const granted = !value

    if (granted) {
      group.dataset.value = 'true'
      group.innerText = 'Y'
    } else {
      group.dataset.value = 'false'
      group.innerText = 'N'
    }

    if (original !== granted) {
      group.dataset.changed = 'true'
      commit.dataset.enabled = 'true'
      rollback.dataset.enabled = 'true'
    } else {
      group.dataset.changed = 'false'
      commit.dataset.enabled = 'false'
      rollback.dataset.enabled = 'false'
    }
  }
}
