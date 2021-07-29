import * as doors from './doors.js'

export function onCommit (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'door':
      doors.commit(row)
      break
  }
}

export function onRollback (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'door':
      doors.rollback(row)
      break
  }
}

export function onRollbackAll (tag, event) {
  switch (tag) {
    case 'door': {
      const rows = document.getElementById('doors').querySelector('table tbody').rows
      for (let i = rows.length; i > 0; i--) {
        doors.rollback(rows[i - 1])
      }
      break
    }
  }
}
