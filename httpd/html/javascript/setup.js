import { postAsForm } from './uhppoted.js'

export function setup (event) {
  dismiss()

  event.preventDefault()

  const credentials = {
    uid: document.getElementById('uid').value,
    pwd: document.getElementById('pwd').value
  }

  postAsForm('/setup', credentials)
    .then(response => {
      console.log('>>', response)
      switch (response.status) {
      //   case 200:
      //     if (response.redirected) {
      //       window.location = response.url
      //     } else {
      //       window.location = '/index.html'
      //     }
      //     return

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

  console.log(msg)

  if (text != null) {
    text.value = `${msg}`
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
