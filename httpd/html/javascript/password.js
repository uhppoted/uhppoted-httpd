/* global messages */

import { postAsForm } from './uhppoted.js'

export function onPassword (event) {
  event.preventDefault()

  dismiss()

  const referrer = document.referrer
  const uid = document.getElementById('uid').value
  const old = document.getElementById('old').value
  const pwd = document.getElementById('pwd').value
  const pwd2 = document.getElementById('pwd2').value

  if (pwd !== pwd2) {
    warning('Passwords do not match')
    return
  }

  const credentials = {
    uid: uid,
    old: old,
    pwd: pwd
  }

  postAsForm('/password', credentials)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            window.location = response.url
          } else {
            window.location = referrer
          }
          return

        case 401:
          throw new Error(messages.unauthorized)

        default:
          return response.text()
      }
    })
    .then(msg => {
      if (msg) {
        throw new Error(msg.trim())
      }
    })
    .catch(function (err) {
      warning(`${err.message}`)
    })
}

export function onEnableOTP (event) {
  event.preventDefault()

  const enable = document.getElementById('otp-enabled')
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')
  const qr = document.getElementById('qrcode')

  getOTP().then((ok) => {
    if (ok) {
      enable.checked = true
      qr.classList.add('visible')
      show.classList.remove('visible')
      hide.classList.add('visible')
    } else {
      enable.checked = false
      qr.classList.remove('visible')
      show.classList.add('visible')
      hide.classList.remove('visible')
    }
  })
}

export function onRevokeOTP (event) {
  event.preventDefault()

  const enable = document.getElementById('otp-enabled')
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')
  const qr = document.getElementById('qrcode')

  revokeOTP().then((ok) => {
    if (ok) {
      enable.checked = false
      enable.disabled = false
      qr.classList.remove('visible')
      show.classList.remove('visible')
      hide.classList.remove('visible')
      warning('OTP revoked')
    }
  })
}

export function onShowOTP (event) {
  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('old').value
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')
  const qr = document.getElementById('qrcode')
  const auth = btoa(`${uid}:${pwd}`)

  URL.revokeObjectURL(qr.src)

  dismiss()

  GET('/otp', `Basic ${auth}`)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            window.location = response.url
            return ''
          } else {
            return response.blob()
          }

        case 401:
          throw new Error(messages.unauthorized)

        default:
          return response
            .text()
            .then(err => { throw new Error(err) })
      }
    })
    .then((v) => {
      if (v instanceof Blob && qr) {
        qr.src = URL.createObjectURL(v)
        qr.classList.add('visible')
        show.classList.remove('visible')
        hide.classList.add('visible')
      }
    })
    .catch(function (err) {
      warning(`${err.message}`)
    })
}

export function onHideOTP (event) {
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')
  const qr = document.getElementById('qrcode')
  const url = qr.src

  show.classList.add('visible')
  hide.classList.remove('visible')
  qr.classList.remove('visible')
  URL.revokeObjectURL(url)
}

export function onVerifyOTP (event) {
  event.preventDefault()

  dismiss()

  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('old').value
  const otp = document.getElementById('otp').value

  const body = {
    uid: uid,
    pwd: pwd,
    otp: otp
  }

  postAsForm('/otp', body)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            window.location = response.url
          }
          return 'OK'

        case 401:
          throw new Error(messages.unauthorized)

        default:
          return response
            .text()
            .then(err => { throw new Error(err) })
      }
    })
    .then((v) => {
      warning('OTP verified and enabled')
    })
    .catch(function (err) {
      warning(`${err.message}`)
    })
}

async function getOTP (event) {
  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('old').value
  const qr = document.getElementById('qrcode')
  const auth = btoa(`${uid}:${pwd}`)

  URL.revokeObjectURL(qr.src)

  dismiss()

  return GET('/otp', `Basic ${auth}`)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            window.location = response.url
            return ''
          } else {
            return response.blob()
          }

        case 401:
          throw new Error(messages.unauthorized)

        default:
          return response
            .text()
            .then(err => { throw new Error(err) })
      }
    })
    .then((v) => {
      if (v instanceof Blob && qr) {
        qr.src = URL.createObjectURL(v)
        return true
      }
    })
    .catch(function (err) {
      warning(`${err.message}`)
      return false
    })
}

async function revokeOTP () {
  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('old').value
  const auth = btoa(`${uid}:${pwd}`)

  dismiss()

  return DELETE('/otp', `Basic ${auth}`)
    .then(response => {
      switch (response.status) {
        case 200:
          console.log(response)
          if (response.redirected) {
            window.location = response.url
            return false
          } else {
            return true
          }

        case 401:
          throw new Error(messages.unauthorized)

        default:
          return response
            .text()
            .then(err => { throw new Error(err) })
      }
    })
    .catch(function (err) {
      warning(`${err.message}`)
      return false
    })
}

async function GET (url = '', authorization = '') {
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
      return response
    })
    .catch(function (err) {
      throw err
    })
}

async function DELETE (url = '', authorization = '') {
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

function warning (msg) {
  const message = document.getElementById('message')
  const text = document.getElementById('warning')

  if (text != null) {
    text.value = msg
    message.classList.add('visible')
  } else {
    alert(msg)
  }
}

function dismiss () {
  const message = document.getElementById('message')
  const text = document.getElementById('warning')

  if (message) {
    message.classList.remove('visible')
  }

  if (text) {
    text.value = ''
  }
}
