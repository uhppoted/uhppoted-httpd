class DoorMode {
  constructor (input, list) {
    this.input = input
    this.list = list

    this.inputHasFocus = false
    this.listHasFocus = false
    this.hasHover = false

    this.options = []
    this.option = null
    this.first = null
    this.last = null

    // ... setup input field
    this.input.addEventListener('keydown', this.onInputKeyDown.bind(this))
    this.input.addEventListener('keyup', this.onInputKeyUp.bind(this))
    this.input.addEventListener('click', this.onInputClick.bind(this))
    this.input.addEventListener('focus', this.onInputFocus.bind(this))
    this.input.addEventListener('blur', this.onInputBlur.bind(this))

    // ... setup options list
    this.list.addEventListener('mouseover', this.onMouseOver.bind(this))
    this.list.addEventListener('mouseout', this.onMouseOut.bind(this))
  }

  initialise () {
    const options = new Set(['controlled', 'normally open', 'normally closed'])

    for (const e of [...(this.list.children)]) {
      this.list.removeChild(e)
    }

    this.options.length = 0
    this.first = null
    this.last = null

    options.forEach(o => {
      const li = document.createElement('li')

      li.appendChild(document.createTextNode(o))
      li.addEventListener('click', this.onOptionClick.bind(this))
      li.addEventListener('mouseover', this.onMouseOver.bind(this))
      li.addEventListener('mouseout', this.onMouseOut.bind(this))

      this.list.appendChild(li)
      this.options.push(li)
    })

    if (this.options.length > 0) {
      this.first = this.options[0]
      this.last = this.options[this.options.length - 1]
    }
  }

  setValue (value) {
    this.input.value = value
  }

  setOption (option) {
    if (option) {
      this.option = option
      this.setCurrentOptionStyle(this.option)
    }
  }

  setFocusDoorMode () {
    this.list.classList.remove('focus')
    this.input.parentNode.classList.add('focus')
    this.inputHasFocus = true
    this.listHasFocus = false
  }

  setVisualFocusListbox () {
    this.input.parentNode.classList.remove('focus')
    this.inputHasFocus = false
    this.listHasFocus = true
    this.list.classList.add('focus')
  }

  removeVisualFocusAll () {
    this.input.parentNode.classList.remove('focus')
    this.inputHasFocus = false
    this.listHasFocus = false
    this.list.classList.remove('focus')
    this.option = null
  }

  setCurrentOptionStyle (option) {
    for (let i = 0; i < this.options.length; i++) {
      const opt = this.options[i]
      if (opt === option) {
        opt.classList.add('selected')
        if (this.list.scrollTop + this.list.offsetHeight < opt.offsetTop + opt.offsetHeight) {
          this.list.scrollTop = opt.offsetTop + opt.offsetHeight - this.list.offsetHeight
        } else if (this.list.scrollTop > opt.offsetTop + 2) {
          this.list.scrollTop = opt.offsetTop
        }
      } else {
        opt.classList.remove('selected')
      }
    }
  }

  getPreviousOption (currentOption) {
    if (currentOption !== this.first) {
      const index = this.options.indexOf(currentOption)
      return this.options[index - 1]
    }
    return this.last
  }

  getNextOption (currentOption) {
    if (currentOption !== this.last) {
      const index = this.options.indexOf(currentOption)
      return this.options[index + 1]
    }

    return this.first
  }

  isOpen () {
    return this.list.style.display === 'block'
  }

  isClosed () {
    return this.list.style.display !== 'block'
  }

  hasOptions () {
    return this.options.length
  }

  open () {
    const rect = this.input.getBoundingClientRect()

    this.list.style.top = `${rect.y + rect.height}px`
    this.list.style.left = `${rect.x}px`
    this.list.style.display = 'block'
  }

  close (force) {
    if (typeof force !== 'boolean') {
      force = false
    }

    if (force || (!this.inputHasFocus && !this.listHasFocus && !this.hasHover)) {
      this.setCurrentOptionStyle(false)
      this.list.style.display = 'none'
    }
  }

  // input field events
  onInputKeyDown (event) {
    let flag = false
    const altKey = event.altKey

    if (event.ctrlKey || event.shiftKey) {
      return
    }

    switch (event.key) {
      case 'Enter':
        if (this.listHasFocus) {
          this.setValue(this.option.textContent)
        }
        this.close(true)
        this.setFocusDoorMode()
        flag = true
        break

      case 'Down':
      case 'ArrowDown':
        if (this.options.length > 0) {
          if (altKey) {
            this.open()
          } else {
            this.open()
            if (this.listHasFocus) {
              this.setOption(this.getNextOption(this.option))
              this.setVisualFocusListbox()
            } else {
              this.setOption(this.first)
              this.setVisualFocusListbox()
            }
          }
        }
        flag = true
        break

      case 'Up':
      case 'ArrowUp':
        if (this.hasOptions()) {
          if (this.listHasFocus) {
            this.setOption(this.getPreviousOption(this.option))
          } else {
            this.open()
            if (!altKey) {
              this.setOption(this.last)
              this.setVisualFocusListbox()
            }
          }
        }
        flag = true
        break

      case 'Esc':
      case 'Escape':
        if (this.isOpen()) {
          this.close(true)
          this.setFocusDoorMode()
        } else {
          this.setValue('')
          this.input.value = ''
        }
        this.option = null
        flag = true
        break

      case 'Tab':
        this.close(true)
        if (this.listHasFocus) {
          if (this.option) {
            this.setValue(this.option.textContent)
          }
        }
        break

      case 'Home':
        this.input.setSelectionRange(0, 0)
        flag = true
        break

      case 'End': {
        const length = this.input.value.length
        this.input.setSelectionRange(length, length)
        flag = true
      }
        break
    }

    if (flag) {
      event.stopPropagation()
      event.preventDefault()
    }
  }

  onInputKeyUp (event) {
    if (event.key === 'Escape' || event.key === 'Esc') {
      return
    }

    switch (event.key) {
      case 'Backspace':
        this.setFocusDoorMode()
        this.setCurrentOptionStyle(false)
        this.option = null

        event.stopPropagation()
        event.preventDefault()
        break

      case 'Left':
      case 'ArrowLeft':
      case 'Right':
      case 'ArrowRight':
      case 'Home':
      case 'End':
        this.option = null
        this.setCurrentOptionStyle(false)
        this.setFocusDoorMode()

        event.stopPropagation()
        event.preventDefault()
        break
    }
  }

  onInputClick () {
    if (this.isOpen()) {
      this.close(true)
    } else {
      this.open()
    }
  }

  onInputFocus () {
    this.setFocusDoorMode()
    this.option = null
    this.setCurrentOptionStyle(null)
  }

  onInputBlur () {
    this.inputHasFocus = false
    this.setCurrentOptionStyle(null)
    this.removeVisualFocusAll()
    setTimeout(this.close.bind(this, false), 100)
  }

  // option events
  onOptionClick (event) {
    this.input.title = event.target.textContent
    this.input.value = event.target.textContent
    this.close(true)

    this.input.dispatchEvent(new Event('change', { bubbles: false, cancelable: true }))
  }

  // mouse events
  onMouseOver () {
    this.hasHover = true
  }

  onMouseOut () {
    this.hasHover = false
    setTimeout(this.close.bind(this, false), 100)
  }
}

export function initialise (mode) {
  const input = mode.querySelector('input')
  const list = mode.querySelector('ul')
  const cb = new DoorMode(input, list)

  cb.initialise()

  return cb
}

export function set (cb, dt) {
  if (dt && !Number.isNaN(dt)) {
    // ??? What to do ??
  }
}
