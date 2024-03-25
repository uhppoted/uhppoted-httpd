/* global messages */

import { GET, POST, DELETE } from './uhppoted.js'

let expired = -1

export function onPassword (event) {
  event.preventDefault()

  dismiss()

  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('pwd').value
  const pwd1 = document.getElementById('pwd1').value
  const pwd2 = document.getElementById('pwd2').value
  const auth = btoa(`${uid}:${pwd}`)

  if (pwd1 !== pwd2) {
    warning('Passwords do not match')
    return
  }

  const credentials = {
    password: pwd1
  }

  POST('/password', `Basic ${auth}`, credentials)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            window.location = response.url
          } else {
            warning('Password changed')
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
  const fieldset = document.querySelector('#OTP fieldset')
  const enable = document.querySelector('#otp-enable')
  const qrcode = document.getElementById('qrcode')

  if (!enable.checked) {
    fieldset.dataset.enabled = 'false'
    return
  }

  getOTP().then((ok) => {
    if (ok) {
      fieldset.dataset.enabled = 'pending'
      fieldset.dataset.otp = 'show'
      qrcode.classList.remove('fadeOut')
      qrcode.classList.add('fadeIn')
    } else {
      fieldset.dataset.enabled = 'false'
      enable.checked = false
    }
  })
}

export function onRevokeOTP (event) {
  event.preventDefault()

  const fieldset = document.querySelector('#OTP fieldset')
  const enable = document.querySelector('#otp-enable')
  const show = document.getElementById('show-otp')
  const hide = document.getElementById('hide-otp')

  revokeOTP().then((ok) => {
    if (ok) {
      fieldset.dataset.enabled = 'false'
      enable.checked = false
      enable.disabled = false
      show.classList.remove('visible')
      hide.classList.remove('visible')
      warning('OTP revoked')
    }
  })
}

export function onHideOTP (event) {
  const fieldset = document.querySelector('#OTP fieldset')
  const qrcode = document.getElementById('qrcode')
  const url = qrcode.src

  fieldset.dataset.otp = 'hide'
  URL.revokeObjectURL(url)

  clearTimeout(expired)
}

export function onShowOTP (event) {
  const fieldset = document.querySelector('#OTP fieldset')
  const qrcode = document.getElementById('qrcode')

  URL.revokeObjectURL(qrcode.src)

  dismiss()

  getOTP().then((ok) => {
    if (ok) {
      qrcode.classList.remove('fadeOut')
      fieldset.dataset.otp = 'show'
    }
  })
}

export function onVerifyOTP (event) {
  event.preventDefault()

  dismiss()

  const fieldset = document.querySelector('#OTP fieldset')
  const checkbox = document.querySelector('#otp-enable')
  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('pwd').value
  const otp = document.getElementById('otp').value
  const auth = btoa(`${uid}:${pwd}`)

  const body = {
    otp: otp
  }

  POST('/otp', `Basic ${auth}`, body)
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
      fieldset.dataset.enabled = 'true'
      checkbox.disabled = true
      warning('OTP verified and enabled')
    })
    .catch(function (err) {
      warning(`${err.message}`)
    })
}

async function getOTP (event) {
  const uid = document.getElementById('uid').value
  const pwd = document.getElementById('pwd').value
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
            let expires = 60000
            for (const [k, v] of response.headers.entries()) {
              if (k.toLowerCase() === 'x-uhppoted-httpd-otp-expires') {
                expires = Number(v) * 1000
              }
            }

            clearTimeout(expired)
            expired = setTimeout(otpExpired, expires)

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
    .then((blob) => {
      if (blob instanceof Blob && qr) {
        qr.src = URL.createObjectURL(blob)
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
  const pwd = document.getElementById('pwd').value
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

function otpExpired () {
  const fieldset = document.querySelector('#OTP fieldset')
  const qrcode = document.getElementById('qrcode')

  if (fieldset.dataset.enabled === 'pending') {
    qrcode.classList.add('fadeOut')
    fieldset.dataset.otp = 'hide'
    URL.revokeObjectURL(qrcode.src)
  }
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
