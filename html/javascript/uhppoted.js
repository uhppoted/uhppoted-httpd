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

export function onEdit (event, value) {
  const element = document.getElementById(event.target.id)

  if (value) {
    element.parentElement.innerHTML = '<input id="' + event.target.id + '" type="checkbox" value="Y" onclick="return onTick(event)" checked>'
  } else {
    element.parentElement.innerHTML = '<input id="' + event.target.id + '" type="checkbox" value="N" onclick="return onTick(event)">'
  }
}

export function onTick (event) {
  console.log('onTick', event.target.checked, event.target.indeterminate)

  // event.target.checked = false
  // event.target.indeterminate = true
  // event.stopPropagation()
  return true
}
