/* global messages */

import { postAsForm, warning } from './uhppoted.js'

export function login (event) {
  event.preventDefault()

  const credentials = {
    uid: document.getElementById('uid').value,
    pwd: document.getElementById('pwd').value
  }

  document.getElementById('message').style.display = 'none'

  postAsForm('/authenticate', credentials)
    .then(response => {
      switch (response.status) {
        case 200:
          if (response.redirected) {
            return response.url
          } else {
            return '/'
          }

        case 401:
          throw new Error(messages.unauthorized)

        default:
          throw new Error(response.text())
      }
    })
    .then(url => {
      window.location = url
    })
    .catch(function (err) {
      warning(err)
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
