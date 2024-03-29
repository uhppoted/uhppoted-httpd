/* global messages */

import { dismiss, postAsForm } from './uhppoted.js'

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
