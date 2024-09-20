let refreshTimer
let idleTimer
let disconnected

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

export function setRefresh (f) {
  refreshTimer = setInterval(f, 15000)
}

export function onIdle () {
  onSignOut()
}

export function onMenu (event, show) {
  if (show) {
    document.querySelector('#user div.menu').style.display = 'block'
  } else {
    document.querySelector('#user div.menu').style.display = 'none'
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
      if (msg) {
        warning(msg)
      }
    })
    .catch(function (err) {
      console.error(err)
      offline()
    })
}

export function onSynchronizeACL (event) {
  if (event != null) {
    event.preventDefault()
  }

  const location = window.location

  postAsJSON('/synchronize/ACL', {})
    .then(response => {
      if (response.status === 200 && response.redirected) {
        window.location = response.url
      } else if (response.status === 200) {
        window.location = location
      } else {
        return response.text()
      }
    })
    .then(msg => {
      warning(msg)
    })
    .catch(function (err) {
      console.error(err)
      offline()
    })
}

export function onSynchronizeDateTime (event) {
  if (event != null) {
    event.preventDefault()
  }

  const location = window.location

  postAsJSON('/synchronize/datetime', {})
    .then(response => {
      if (response.status === 200 && response.redirected) {
        window.location = response.url
      } else if (response.status === 200) {
        window.location = location
      } else {
        return response.text()
      }
    })
    .then(msg => {
      warning(msg)
    })
    .catch(function (err) {
      console.error(err)
      offline()
    })
}

export function onSynchronizeDoors (event) {
  console.log('onSynchronizeDoors')
  if (event != null) {
    event.preventDefault()
  }

  const location = window.location

  postAsJSON('/synchronize/doors', {})
    .then(response => {
      console.error('> 1', response)
      if (response.status === 200 && response.redirected) {
        window.location = response.url
      } else if (response.status === 200) {
        window.location = location
      } else {
        return response.text()
      }
    })
    .then(msg => {
      warning(msg)
    })
    .catch(function (err) {
      console.error(err)
      offline()
    })
}

export function onShowHidePassword (event, id) {
  const pwd = document.getElementById(id)
  const eye = event.target

  if (pwd.type === 'password') {
    pwd.type = 'text'
    eye.src = eye.dataset.hide
  } else {
    pwd.type = 'password'
    eye.src = eye.dataset.show
  }
}

export function retheme (theme) {
  const expires = new Date()
  const stylesheets = document.querySelectorAll("link[rel='stylesheet']")
  const images = document.querySelectorAll('img')

  expires.setFullYear(expires.getFullYear() + 1)

  document.cookie = 'uhppoted-settings=theme:' + theme + '; expires=' + expires.toUTCString()

  stylesheets.forEach(link => {
    const re = new RegExp('(.+?/css)/(.+?)/(.+)', 'i') // eslint-disable-line prefer-regex-literals

    if (re.test(link.href)) {
      const match = link.href.match(re)

      link.href = match[1] + '/' + theme + '/' + match[3]
    }
  })

  images.forEach(img => {
    const re = new RegExp('(.+?/images)/(.+?)/(.+)', 'i') // eslint-disable-line prefer-regex-literals

    if (re.test(img.src)) {
      const match = img.src.match(re)

      img.src = match[1] + '/' + theme + '/' + match[3]
    }
  })
}

export function warning (msg) {
  const message = document.getElementById('message')
  const text = document.getElementById('warning')

  if (message && text) {
    if (msg) {
      message.classList.add('visible')
      text.value = msg
    } else {
      message.classList.remove('visible')
      text.value = ''
    }
  } else if (msg) {
    alert(msg)
  }
}

export function dismiss () {
  const message = document.getElementById('message')
  const text = document.getElementById('warning')

  if (message) {
    message.classList.remove('visible')
  }

  if (text) {
    text.innerText = 'msg' // FIXME - think this was a hack for a layout issue - fix in CSS rather
  }
}

export async function GET (url = '', authorization = '') {
  const init = {
    method: 'GET',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    headers: { Authorization: authorization }
  }

  return await fetch(url, init)
    .then(response => {
      connected(true)
      return response
    })
    .catch(function (err) {
      connected(false)
      throw err
    })
}

export async function POST (url = '', authorization = '', data = {}) {
  dismiss()

  const body = Object.entries(data)
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join('&')
    .replace(/%20/g, '+')

  const init = {
    method: 'POST',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: {
      Authorization: authorization,
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    body: body
  }

  return await fetch(url, init)
    .then(response => {
      return response
    })
    .catch(function (err) {
      throw err
    })
}

export async function DELETE (url = '', authorization = '') {
  const init = {
    method: 'DELETE',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    headers: { Authorization: authorization }
  }

  return await fetch(url, init)
    .then(response => {
      return response
    })
    .catch(function (err) {
      throw err
    })
}

export async function getAsJSON (url = '') {
  const init = {
    method: 'GET',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    redirect: 'follow',
    referrerPolicy: 'no-referrer'
  }

  return await fetch(url, init)
    .then(response => {
      connected(true)
      return response
    })
    .catch(function (err) {
      connected(false)
      throw err
    })
}

export async function postAsForm (url = '', data = {}) {
  dismiss()

  const pairs = []
  for (const name in data) {
    pairs.push(encodeURIComponent(name) + '=' + encodeURIComponent(data[name]))
  }

  const init = {
    method: 'POST',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    body: pairs.join('&').replace(/%20/g, '+')
  }

  return await fetch(url, init)
    .then(response => {
      connected(true)
      return response
    })
    .catch(function (err) {
      connected(false)
      throw err
    })
}

export async function postAsJSON (url = '', data = {}) {
  dismiss()

  const init = {
    method: 'POST',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json' },
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    body: JSON.stringify(data)
  }

  return await fetch(url, init)
    .then(response => {
      connected(true)
      return response
    })
    .catch(function (err) {
      connected(false)
      throw err
    })
}

export function resetIdle () {
  if (idleTimer != null) {
    clearTimeout(idleTimer)
  }

  idleTimer = setTimeout(onIdle, 5 * 60 * 1000)
}

export function busy () {
  const windmill = document.getElementById('windmill')

  if (windmill) {
    const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

    windmill.dataset.count = (queued + 1).toString()
  }
}

export function unbusy () {
  const windmill = document.getElementById('windmill')

  if (windmill) {
    const queued = Math.max(0, (windmill.dataset.count && parseInt(windmill.dataset.count)) | 0)

    if (queued > 1) {
      windmill.dataset.count = (queued - 1).toString()
    } else {
      delete (windmill.dataset.count)
    }
  }
}

export function loading () {
  const windmill = document.getElementById('loading')
  const container = document.getElementById('container')

  if (windmill) {
    windmill.classList.add('visible')
  }

  if (container) {
    container.classList.add('loading')
  }
}

export function loaded () {
  const windmill = document.getElementById('loading')
  const container = document.getElementById('container')

  if (windmill) {
    windmill.classList.remove('visible')
  }

  if (container) {
    container.classList.remove('loading')
  }
}

export function onReload () {
  const message = document.querySelector('#offline + div > p')

  message.innerHTML = '.... trying ....'

  fetch('/index.html', {
    method: 'HEAD',
    mode: 'cors',
    cache: 'no-cache',
    credentials: 'same-origin',
    redirect: 'follow',
    referrerPolicy: 'no-referrer'
  }).then(response => {
    window.location = '/index.html'
  }).catch(function (err) {
    console.error(err)
    message.innerHTML = '(still offline)'
  })
}

function connected (ok) {
  const element = document.querySelector('header #disconnected')
  let monitor = false

  if (window.constants && window.constants.mode && window.constants.mode === 'monitor') {
    monitor = true
  }

  if (element) {
    if (!ok) {
      element.classList.add('visible')
      element.innerHTML = 'disconnected'
    } else if (ok && monitor) {
      element.classList.add('visible')
      element.innerHTML = 'monitor'
    } else {
      element.classList.remove('visible')
    }
  }

  if (ok) {
    disconnected = null
  } else if (!disconnected) {
    disconnected = new Date()
  } else {
    const duration = (new Date() - disconnected)
    const seconds = Math.floor(Math.max(0, Math.round(duration / 1000)))

    if (seconds > 60) {
      console.error(`Disconnected for ${seconds}s`)
      offline()
    }
  }
}

function offline () {
  const cookies = document.cookie.split(';')

  for (let i = 0; i < cookies.length; i++) {
    const cookie = cookies[i]
    const ix = cookie.indexOf('=')
    const name = ix > -1 ? cookie.substr(0, ix) : cookie

    if (name === 'JSESSIONID') {
      document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:00 GMT'
    }
  }

  document.body.innerHTML = '<div id="offline"><div><div><p>SYSTEM OFFLINE</p></div><div><a onclick="onReload()">RELOAD</a></div></div></div><div><p/></div>'

  clearInterval(refreshTimer)
}
