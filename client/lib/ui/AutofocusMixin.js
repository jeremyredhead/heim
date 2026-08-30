import ReactDOM from 'react-dom'

export default {
  componentDidMount() {
    this.checkAutofocus()
  },

  componentDidUpdate() {
    this.checkAutofocus()
  },

  checkAutofocus() {
    const step = this.state.flow.step
    if (step === this._prevFlowStep) {
      return
    }
    this._prevFlowStep = step
    this.applyAutofocus()
  },

  applyAutofocus() {
    const child = ReactDOM.findDOMNode(this).querySelector('[tabindex="1"]')
    if (child) {
      child.focus()
    }
  },
}
