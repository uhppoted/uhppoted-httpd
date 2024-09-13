export class Cache {
  constructor (options = {}) {
    const { elements = true, modified = true } = options

    if (elements) {
      this.elements = new Map()
    }

    if (modified) {
      this.modified = new Map()
    }
  }

  put (oid, field) {
    const key = `${oid}`

    this.elements.set(key, field)
  }

  query (oid) {
    const key = `${oid}`

    if (this.elements) {
      if (!this.elements.has(key)) {
        const e = document.querySelector(`[data-oid="${oid}"]`)

        this.elements.set(key, e)
      }

      return this.elements.get(key)
    }

    return document.querySelector(`[data-oid="${oid}"]`)
  }

  queryModified (oid) {
    const key = `${oid}`

    if (this.modified) {
      if (!this.modified.has(key)) {
        this.modified.set(key, document.querySelectorAll(`[data-oid^="${oid}."].modified`))
      }

      return this.modified.get(key)
    }

    return document.querySelectorAll(`[data-oid^="${oid}."].modified`)
  }
}
