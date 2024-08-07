// /* global messages */
//
// import { postAsForm } from './uhppoted.js'

export function setup (event) {
  dismiss()

  event.preventDefault()

  console.log('>>>>>>>>>>>>>> SETUP')

  // preauth()
  //   .then(response => {
  //     switch (response.status) {
  //       case 200:
  //         return true
//
  //       default:
  //         throw new Error(response.statusText)
  //     }
  //   })
  //   .then(v => {
  //     auth()
  //   })
  //   .catch(function (err) {
  //     warning(`Error logging in (${err.message.toLowerCase()})`)
  //   })
}

// // HEAD request to refresh the uhppoted-httpd-login cookie.
// // (preempts the double login needed if the cookie has expired)
// async function preauth () {
//   const init = {
//     method: 'HEAD',
//     mode: 'cors',
//     cache: 'no-cache',
//     credentials: 'same-origin',
//     redirect: 'follow',
//     referrerPolicy: 'no-referrer'
//   }
//
//   return await fetch('/authenticate', init)
//     .then(response => {
//       return response
//     })
// }

// function auth () {
//   const credentials = {
//     uid: document.getElementById('uid').value,
//     pwd: document.getElementById('pwd').value
//   }
//
//   postAsForm('/authenticate', credentials)
//     .then(response => {
//       switch (response.status) {
//         case 200:
//           if (response.redirected) {
//             window.location = response.url
//           } else {
//             window.location = '/index.html'
//           }
//           return
//
//         case 401:
//           throw new Error(messages.unauthorized)
//
//         default:
//           return response.text()
//       }
//     })
//     .then(msg => {
//       if (msg) {
//         throw new Error(msg.trim())
//       }
//     })
//     .catch(function (err) {
//       warning(`${err.message}`)
//     })
// }

// function warning (msg) {
//   const message = document.getElementById('message')
//   const text = document.getElementById('warning')
//
//   console.log(msg)
//   console.log(text)
//
//   if (text != null) {
//     text.value = `${msg}`
//     message.classList.add('visible')
//   } else {
//     alert(msg)
//   }
// }

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
