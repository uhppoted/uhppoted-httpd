import * as doors from './doors.js'

export function onEdited (tag, event) {
  switch (tag) {
    case 'door':
      doors.set(event.target, event.target.value)
      break
  }
}

export function onEnter (tag, event) {
  if (event.key === 'Enter') {
    switch (tag) {
      case 'door': {
        doors.set(event.target, event.target.value)
        break
      }
    }
  }
}

export function onTick (tag, event) {
  switch (tag) {
    case 'door': {
      doors.set(event.target, event.target.checked)
      break
    }
  }
}

export function onCommit (tag, event) {
  const id = event.target.dataset.record
  const row = document.getElementById(id)

  switch (tag) {
    case 'door':
      doors.commit(row)
      break
  }
}

export function onCommitAll (tag, event, table) {
  const tbody = document.getElementById(table).querySelector('table tbody')
  const rows = tbody.rows
  const list = []

  for (let i = 0; i < rows.length; i++) {
    const row = rows[i]
    if (row.classList.contains('modified') || row.classList.contains('new')) {
      list.push(row)
    }
  }

  switch (tag) {
    case 'door': {
      doors.commit(...list)
      break
    }
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

export function onNew (tag, event) {
  switch (tag) {
    case 'door':
      doors.create()
      break
  }
}
