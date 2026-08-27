import ReactDOM from 'react-dom'

class Box {
  constructor(value = null) {
    this.value = value
  }

  update(newValue) {
    if (newValue === this.value) {
      return false
    }
    this.value = newValue
    return true
  }
}

export default {
  getInitialState() {
    return {prevFlowStep: new Box()}
  },

  componentDidMount() {
    this.checkAutofocus()
  },

  componentDidUpdate() {
    this.checkAutofocus()
  },

  checkAutofocus() {
    const step = this.state.flow.step
    if (!this.state.prevFlowStep.update(step)) {
      return
    }
    this.applyAutofocus()
  },

  applyAutofocus() {
    const child = ReactDOM.findDOMNode(this).querySelector('[tabindex="1"]')
    if (child) {
      child.focus()
    }
  },
}
