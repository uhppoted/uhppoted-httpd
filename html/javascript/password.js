/* global messages */

import { dismiss, postAsForm } from './uhppoted.js'

export function onPassword (event) {
  dismiss()

  event.preventDefault()

  const referrer = document.referrer

  const credentials = {
    uid: document.getElementById('uid').value,
    old: document.getElementById('old').value,
    pwd: document.getElementById('pwd').value,
    pwd2: document.getElementById('pwd2').value
  }

  postAsForm('/password', credentials)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            return { redirected: true, url: response.url }
          } else {
            return { ok: true }
          }

        case 401:
          throw new Error(messages.unauthorized)

        default:
          throw new Error(response.text())
      }
    })
    .then(o => {
      if (o.ok) {
        window.location = referrer
      } else if (o.redirected && o.url) {
        window.location = o.url
      } else {
        throw new Error('system error')
      }
    })
    .catch(function (err) {
      warning(`Error changing password (${err.message.toLowerCase()})`)
    })
}

export function showHidePassword () {
  const pwd = document.getElementById('pwd')
  const eye = document.getElementById('eye')

  if (pwd.type === 'password') {
    pwd.type = 'text'
    eye.src = 'images/eye-slash-solid.svg'
  } else {
    pwd.type = 'password'
    eye.src = 'images/eye-solid.svg'
  }
}

function warning (msg) {
  const message = document.getElementById('message')
  const text = document.getElementById('warning')

  if (text != null) {
    text.innerText = msg
    message.style.visibility = 'visible'
  } else {
    alert(msg)
  }
}
