import { mark, unmark } from './tabular.js'
import { DB } from './db.js'
import { schema } from './schema.js'

export function refreshed () {
  DB.interfaces.forEach(c => {
    updateFromDB(c.OID, c)
  })
}

function updateFromDB (oid, record) {
  const section = document.querySelector(`[data-oid="${oid}"]`)

  if (section) {
    const name = section.querySelector(`[data-oid="${oid}${schema.interfaces.name}"]`)
    const bind = section.querySelector(`[data-oid="${oid}${schema.interfaces.bind}"]`)
    const broadcast = section.querySelector(`[data-oid="${oid}${schema.interfaces.broadcast}"]`)

    update(name, record.name)
    update(bind, record.bind)
    update(broadcast, record.broadcast)
  }
}

export function set (element, value, status) {
  const oid = element.dataset.oid
  const original = element.dataset.original
  const v = value.toString()
  const parent = element.parentElement

  element.dataset.value = v

  if (v !== original) {
    mark('modified', element, parent)
  } else {
    unmark('modified', element, parent)
  }

  percolate(oid)
}

export function rollback (tag, element) {
  const section = document.getElementById(tag)
  const oid = section.dataset.oid

  const children = section.querySelectorAll(`[data-oid^="${oid}."]`)
  children.forEach(e => {
    const parent = e.parentElement

    e.dataset.value = e.dataset.original
    e.value = e.dataset.original
    e.classList.remove('modified')

    parent.classList.remove('modified')
    parent.classList.remove('pending')
  })

  section.classList.remove('modified')
}

export function changeset (element) {
  const section = document.getElementById('interface')
  const oid = section.dataset.oid
  const list = []

  const children = section.querySelectorAll(`[data-oid^="${oid}."]`)
  children.forEach(e => {
    if (e.dataset.value !== e.dataset.original) {
      list.push(e)
    }
  })

  return {
    updated: list,
    deleted: []
  }
}

function modified (oid) {
  const container = document.querySelector(`[data-oid="${oid}"]`)
  let changed = false

  if (container) {
    const list = document.querySelectorAll(`[data-oid^="${oid}."]`)
    list.forEach(e => {
      changed = changed || e.classList.contains('modified')
    })

    if (changed) {
      container.classList.add('modified')
    } else {
      container.classList.remove('modified')
    }
  }
}

function update (element, value, status) {
  if (element) {
    const v = value.toString()
    const oid = element.dataset.oid
    const td = element.parentElement
    const previous = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently edited fields
    if (element.classList.contains('modified')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, td)
      } else if (element.dataset.value !== v) {
        unmark('conflict', element, td)
      } else {
        unmark('conflict', element, td)
        unmark('modified', element, td)
      }

      percolate(oid)
      return
    }

    // check for conflicts with concurrently submitted fields
    if (element.classList.contains('pending')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, td)
      } else {
        unmark('conflict', element, td)
      }

      return
    }

    // update fields not pending or modified
    if (element !== document.activeElement) {
      element.value = v
    }

    set(element, value)
  }
}

function percolate (oid) {
  let oidx = oid
  while (oidx) {
    const match = /(.*?)(?:[.][0-9]+)$/.exec(oidx)
    oidx = match ? match[1] : null
    if (oidx) {
      modified(oidx)
    }
  }
}
