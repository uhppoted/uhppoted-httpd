class Combobox {
  constructor (input, button, list) {
    this.input = input
    this.buttonNode = button
    this.listboxNode = list

    this.comboboxHasVisualFocus = false
    this.listboxHasVisualFocus = false
    this.hasHover = false

    this.allOptions = []
    this.option = null
    this.firstOption = null
    this.lastOption = null
    this.filteredOptions = []

    // ... setup input field
    this.input.addEventListener('keydown', this.onComboboxKeyDown.bind(this))
    this.input.addEventListener('keyup', this.onComboboxKeyUp.bind(this))
    this.input.addEventListener('click', this.onComboboxClick.bind(this))
    this.input.addEventListener('focus', this.onComboboxFocus.bind(this))
    this.input.addEventListener('blur', this.onComboboxBlur.bind(this))

    // ... setup options list
    this.listboxNode.addEventListener('mouseover', this.onListboxMouseover.bind(this))
    this.listboxNode.addEventListener('mouseout', this.onListboxMouseout.bind(this))

    const nodes = this.listboxNode.getElementsByTagName('LI')
    for (let i=0; i<nodes.length; i++) {
      const option = nodes[i]

      this.allOptions.push(option)

      option.addEventListener('click', this.onOptionClick.bind(this))
      option.addEventListener('mouseover', this.onOptionMouseover.bind(this))
      option.addEventListener('mouseout', this.onOptionMouseout.bind(this))
    }

    // ... attach onclick handler to 'open' button
    this.buttonNode.addEventListener('click', this.onButtonClick.bind(this))
  }

  initialise () {
  }

  setActiveDescendant (option) {
    if (option && this.listboxHasVisualFocus) {
      this.input.setAttribute('aria-activedescendant', option.id)
    } else {
      this.input.setAttribute('aria-activedescendant', '')
    }
  }

  setValue (value) {
    this.input.value = value
  }

  setOption (option, flag) {
    if (typeof flag !== 'boolean') {
      flag = false
    }

    if (option) {
      this.option = option
      this.setCurrentOptionStyle(this.option)
      this.setActiveDescendant(this.option)
    }
  }

  setVisualFocusCombobox () {
    this.listboxNode.classList.remove('focus')
    this.input.parentNode.classList.add('focus') // set the focus class to the parent for easier styling
    this.comboboxHasVisualFocus = true
    this.listboxHasVisualFocus = false
    this.setActiveDescendant(false)
  }

  setVisualFocusListbox () {
    this.input.parentNode.classList.remove('focus')
    this.comboboxHasVisualFocus = false
    this.listboxHasVisualFocus = true
    this.listboxNode.classList.add('focus')
    this.setActiveDescendant(this.option)
  }

  removeVisualFocusAll () {
    this.input.parentNode.classList.remove('focus')
    this.comboboxHasVisualFocus = false
    this.listboxHasVisualFocus = false
    this.listboxNode.classList.remove('focus')
    this.option = null
    this.setActiveDescendant(false)
  }

  // ComboboxAutocomplete Events

  setCurrentOptionStyle (option) {
    for (let i = 0; i < this.filteredOptions.length; i++) {
      const opt = this.filteredOptions[i]
      if (opt === option) {
        opt.setAttribute('aria-selected', 'true')
        if (this.listboxNode.scrollTop + this.listboxNode.offsetHeight < opt.offsetTop + opt.offsetHeight) {
          this.listboxNode.scrollTop = opt.offsetTop + opt.offsetHeight - this.listboxNode.offsetHeight
        } else if (this.listboxNode.scrollTop > opt.offsetTop + 2) {
          this.listboxNode.scrollTop = opt.offsetTop
        }
      } else {
        opt.removeAttribute('aria-selected')
      }
    }
  }

  getPreviousOption (currentOption) {
    if (currentOption !== this.firstOption) {
      const index = this.filteredOptions.indexOf(currentOption)
      return this.filteredOptions[index - 1]
    }
    return this.lastOption
  }

  getNextOption (currentOption) {
    if (currentOption !== this.lastOption) {
      const index = this.filteredOptions.indexOf(currentOption)
      return this.filteredOptions[index + 1]
    }
    return this.firstOption
  }

  // list display functions

  doesOptionHaveFocus () {
    return this.input.getAttribute('aria-activedescendant') !== ''
  }

  isOpen () {
    return this.listboxNode.style.display === 'block'
  }

  isClosed () {
    return this.listboxNode.style.display !== 'block'
  }

  hasOptions () {
    return this.filteredOptions.length
  }

  open () {
    this.listboxNode.style.display = 'block'
    this.input.setAttribute('aria-expanded', 'true')
    this.buttonNode.setAttribute('aria-expanded', 'true')
  }

  close (force) {
    if (typeof force !== 'boolean') {
      force = false
    }

    if (force || (!this.comboboxHasVisualFocus && !this.listboxHasVisualFocus && !this.hasHover)) {
      this.setCurrentOptionStyle(false)
      this.listboxNode.style.display = 'none'
      this.input.setAttribute('aria-expanded', 'false')
      this.buttonNode.setAttribute('aria-expanded', 'false')
      this.setActiveDescendant(false)
    }
  }

  /* combobox Events */

  onComboboxKeyDown (event) {
    let flag = false
    const altKey = event.altKey

    if (event.ctrlKey || event.shiftKey) {
      return
    }

    switch (event.key) {
      case 'Enter':
        if (this.listboxHasVisualFocus) {
          this.setValue(this.option.textContent)
        }
        this.close(true)
        this.setVisualFocusCombobox()
        flag = true
        break

      case 'Down':
      case 'ArrowDown':
        if (this.filteredOptions.length > 0) {
          if (altKey) {
            this.open()
          } else {
            this.open()
            if (this.listboxHasVisualFocus) {
              this.setOption(this.getNextOption(this.option), true)
              this.setVisualFocusListbox()
            } else {
              this.setOption(this.firstOption, true)
              this.setVisualFocusListbox()
            }
          }
        }
        flag = true
        break

      case 'Up':
      case 'ArrowUp':
        if (this.hasOptions()) {
          if (this.listboxHasVisualFocus) {
            this.setOption(this.getPreviousOption(this.option), true)
          } else {
            this.open()
            if (!altKey) {
              this.setOption(this.lastOption, true)
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
          this.setVisualFocusCombobox()
        } else {
          this.setValue('')
          this.input.value = ''
        }
        this.option = null
        flag = true
        break

      case 'Tab':
        this.close(true)
        if (this.listboxHasVisualFocus) {
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

      default:
        break
    }

    if (flag) {
      event.stopPropagation()
      event.preventDefault()
    }
  }

  onComboboxKeyUp (event) {
    if (event.key === 'Escape' || event.key === 'Esc') {
      return
    }

    switch (event.key) {
      case 'Backspace':
        this.setVisualFocusCombobox()
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
        this.setVisualFocusCombobox()

        event.stopPropagation()
        event.preventDefault()
        break
    }
  }

  onComboboxClick () {
    if (this.isOpen()) {
      this.close(true)
    } else {
      this.open()
    }
  }

  onComboboxFocus () {
    this.setVisualFocusCombobox()
    this.option = null
    this.setCurrentOptionStyle(null)
  }

  onComboboxBlur () {
    this.comboboxHasVisualFocus = false
    this.setCurrentOptionStyle(null)
    this.removeVisualFocusAll()
    setTimeout(this.close.bind(this, false), 300)
  }

  onButtonClick () {
    if (this.isOpen()) {
      this.close(true)
    } else {
      this.open()
    }

    this.input.focus()
    this.setVisualFocusCombobox()
  }

  // list events
  onListboxMouseover () {
    this.hasHover = true
  }

  onListboxMouseout () {
    this.hasHover = false
    setTimeout(this.close.bind(this, false), 300)
  }

  // option events
  onOptionClick (event) {
    this.input.title = event.target.textContent
    this.input.value = event.target.textContent
    this.close(true)

    this.input.dispatchEvent(new Event('change', { bubbles: false, cancelable: true }))
  }

  onOptionMouseover () {
    this.hasHover = true
    this.open()
  }

  onOptionMouseout () {
    this.hasHover = false
    setTimeout(this.close.bind(this, false), 300)
  }
}

export function initialise (combobox) {
  const input = combobox.querySelector('input')
  const button = combobox.querySelector('button')
  const list = combobox.querySelector('ul')
  const cb = new Combobox(input, button, list)

  cb.initialise()
}
