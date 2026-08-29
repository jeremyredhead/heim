import '@babel/polyfill'
import 'whatwg-fetch'

import clientRoom from './clientRoom'
import clientVerifyEmail from './clientVerifyEmail'
import clientResetPassword from './clientResetPassword'

// setup globals (used by env frame)
window.uiwindow = window.top
window.uidocument = window.top.document

const tag = document.getElementById('heim-js')
const entrypoint = tag.getAttribute('data-entrypoint')
if (!entrypoint) {
  clientRoom()
} else {
  const crashHandler = require('./ui/crashHandler').default
  document.addEventListener('ravenHandle', crashHandler)
  if (entrypoint === 'verify-email') {
    clientVerifyEmail()
  } else if (entrypoint === 'reset-password') {
    clientResetPassword()
  }
}
