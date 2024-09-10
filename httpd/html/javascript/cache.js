export class Cache {
  constructor () {
    this.elements = new Map()
    this.modified = new Map()
  }

  query (oid) {
    const key = `${oid}`

    if (!this.elements.has(key)) {
      this.elements.set(key, document.querySelector(`[data-oid="${oid}"]`))
    }

    return this.elements.get(key)
  }

  queryModified (oid) {
    const key = `${oid}`

    if (!this.modified.has(key)) {
      this.modified.set(key, document.querySelectorAll(`[data-oid^="${oid}."].modified`))
    }

    return this.modified.get(key)
  }
}
