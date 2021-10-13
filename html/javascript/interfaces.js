import { set } from './tabular.js'
import { DB } from './db.js'

export function refreshed () {
  DB.interfaces.forEach(c => {
    updateFromDB(c.OID, c)
  })
}

function updateFromDB (oid, record) {
  const section = document.querySelector(`[data-oid="${oid}"]`)

  if (section) {
    const name = section.querySelector(`[data-oid="${oid}.1"]`)
    const bind = section.querySelector(`[data-oid="${oid}.2"]`)
    const broadcast = section.querySelector(`[data-oid="${oid}.3"]`)
    const listen = section.querySelector(`[data-oid="${oid}.4"]`)

    update(name, record.name)
    update(bind, record.bind)
    update(broadcast, record.broadcast)
    update(listen, record.listen)
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
    const flag = document.getElementById(`F${oid}`)
    const previous = element.dataset.original

    element.dataset.original = v

    // check for conflicts with concurrently edited fields
    if (element.classList.contains('modified')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, flag)
      } else if (element.dataset.value !== v) {
        unmark('conflict', element, flag)
      } else {
        unmark('conflict', element, flag)
        unmark('modified', element, flag)
      }

      percolate(oid)
      return
    }

    // check for conflicts with concurrently submitted fields
    if (element.classList.contains('pending')) {
      if (previous !== v && element.dataset.value !== v) {
        mark('conflict', element, flag)
      } else {
        unmark('conflict', element, flag)
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

function mark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.add(clazz)
    }
  })
}

function unmark (clazz, ...elements) {
  elements.forEach(e => {
    if (e) {
      e.classList.remove(clazz)
    }
  })
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
