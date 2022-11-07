/* global messages */

import { postAsForm, dismiss } from './uhppoted.js'

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

export function onShowOTP (event) {
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')
  const qr = document.getElementById('qrcode')

  URL.revokeObjectURL(qr.src)

  get('/otp')
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
      if (v instanceof Blob && qrcode) {
        qrcode.src = URL.createObjectURL(v)
        qrcode.classList.add('visible')
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

export function onOTP (event) {
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

async function get (url = '') {
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
    text.innerText = msg
    message.classList.add('visible')
  } else {
    alert(msg)
  }
}
