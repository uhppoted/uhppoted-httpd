import { timezones } from './timezones.js'

class Combobox {
  constructor (input, list) {
    this.input = input
    this.list = list

    this.inputHasFocus = false
    this.listHasFocus = false
    this.hasHover = false

    this.allOptions = []
    this.filteredOptions = []
    this.option = null
    this.firstOption = null
    this.lastOption = null

    // ... setup input field
    this.input.addEventListener('keydown', this.onInputKeyDown.bind(this))
    this.input.addEventListener('keyup', this.onInputKeyUp.bind(this))
    this.input.addEventListener('click', this.onInputClick.bind(this))
    this.input.addEventListener('focus', this.onInputFocus.bind(this))
    this.input.addEventListener('blur', this.onInputBlur.bind(this))

    // ... setup options list
    this.list.addEventListener('mouseover', this.onMouseOver.bind(this))
    this.list.addEventListener('mouseout', this.onMouseOut.bind(this))

    const now = new Date()
    timezones.forEach(tz => {
      const text = now.toLocaleString('default', { timeZone: tz }) + ' ' + tz
      const li = document.createElement('li')

      li.appendChild(document.createTextNode(text))
      li.addEventListener('click', this.onOptionClick.bind(this))
      li.addEventListener('mouseover', this.onMouseOver.bind(this))
      li.addEventListener('mouseout', this.onMouseOut.bind(this))

      this.allOptions.push(li)
      this.list.appendChild(li)
    })
  }

  initialise () {
  }

  setActiveDescendant (option) {
    if (option && this.listHasFocus) {
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

  setFocusCombobox () {
    this.list.classList.remove('focus')
    this.input.parentNode.classList.add('focus') // set the focus class to the parent for easier styling
    this.inputHasFocus = true
    this.listHasFocus = false
    this.setActiveDescendant(false)
  }

  setVisualFocusListbox () {
    this.input.parentNode.classList.remove('focus')
    this.inputHasFocus = false
    this.listHasFocus = true
    this.list.classList.add('focus')
    this.setActiveDescendant(this.option)
  }

  removeVisualFocusAll () {
    this.input.parentNode.classList.remove('focus')
    this.inputHasFocus = false
    this.listHasFocus = false
    this.list.classList.remove('focus')
    this.option = null
    this.setActiveDescendant(false)
  }

  // autocomplete Events
  setCurrentOptionStyle (option) {
    for (let i = 0; i < this.filteredOptions.length; i++) {
      const opt = this.filteredOptions[i]
      if (opt === option) {
        opt.setAttribute('aria-selected', 'true')
        if (this.list.scrollTop + this.list.offsetHeight < opt.offsetTop + opt.offsetHeight) {
          this.list.scrollTop = opt.offsetTop + opt.offsetHeight - this.list.offsetHeight
        } else if (this.list.scrollTop > opt.offsetTop + 2) {
          this.list.scrollTop = opt.offsetTop
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
    return this.list.style.display === 'block'
  }

  isClosed () {
    return this.list.style.display !== 'block'
  }

  hasOptions () {
    return this.filteredOptions.length
  }

  open () {
    const rect = this.input.getBoundingClientRect()

    this.list.style.top = `${rect.y+rect.height}px`
    this.list.style.left = `${rect.x}px`
    this.list.style.display = 'block'

    this.input.setAttribute('aria-expanded', 'true')
  }

  close (force) {
    if (typeof force !== 'boolean') {
      force = false
    }

    if (force || (!this.inputHasFocus && !this.listHasFocus && !this.hasHover)) {
      this.setCurrentOptionStyle(false)
      this.list.style.display = 'none'
      this.input.setAttribute('aria-expanded', 'false')
      this.setActiveDescendant(false)
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
        this.setFocusCombobox()
        flag = true
        break

      case 'Down':
      case 'ArrowDown':
        if (this.filteredOptions.length > 0) {
          if (altKey) {
            this.open()
          } else {
            this.open()
            if (this.listHasFocus) {
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
          if (this.listHasFocus) {
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
          this.setFocusCombobox()
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
        this.setFocusCombobox()
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
        this.setFocusCombobox()

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
    this.setFocusCombobox()
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

export function initialise (combobox) {
  const input = combobox.querySelector('input')
  const list = combobox.querySelector('ul')
  const cb = new Combobox(input, list)

  cb.initialise()

  // console.log(new Date().toLocaleString('default', { timeZone: 'UTC' }));
  // console.log(new Date().toLocaleString('default', { timeZone: 'PST' }));
  // console.log(new Date().toLocaleString('default', { timeZone: 'GMT' }));
  // console.log(new Date().toLocaleString('default', { timeZone: 'Africa/Maseru' }));
  // console.log(new Date().toLocaleString('default', { timeZone: 'Etc/GMT-2' }));
  // console.log(new Date().toLocaleString('default', { timeZone: 'PST8PDT' }));
}

// function format (timestamp) {
//   const dt = Date.parse(timestamp)
//   const fmt = function (v) {
//     return v < 10 ? '0' + v.toString() : v.toString()
//   }
//
//   if (!isNaN(dt)) {
//     const date = new Date(dt)
//     const year = date.getFullYear()
//     const month = fmt(date.getMonth() + 1)
//     const day = fmt(date.getDate())
//     const hour = fmt(date.getHours())
//     const minute = fmt(date.getMinutes())
//     const second = fmt(date.getSeconds())
//
//     return `${year}-${month}-${day} ${hour}:${minute}:${second}`
//   }
//
//   return ''
// }
