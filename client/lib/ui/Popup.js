import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'
import classNames from 'classnames'

module.exports = createReactClass({
  displayName: 'Popup',

  propTypes: {
    kind: PropTypes.string,
    className: PropTypes.string,
    onDismiss: PropTypes.func,
    children: PropTypes.node,
  },

  componentDidMount() {
    if (this.props.kind === 'native') {
      ReactDOM.findDOMNode(this).show()
    } else if (this.props.kind === 'native-modal') {
      ReactDOM.findDOMNode(this).showModal()
    }
    setImmediate(() => {
      Heim.addEventListener(uidocument, Heim.isTouch ? 'touchstart' : 'click', this.onOutsideClick, false)
    })
  },

  componentWillUnmount() {
    Heim.removeEventListener(uidocument, Heim.isTouch ? 'touchstart' : 'click', this.onOutsideClick, false)
  },

  onOutsideClick(ev) {
    if (!ReactDOM.findDOMNode(this).contains(ev.target) && this.props.onDismiss) {
      this.props.onDismiss(ev)
    }
  },

  render() {
    const isNative = this.props.kind === 'native' || this.props.kind === 'native-modal'
    return React.createElement(
      isNative ? 'dialog' : 'div',
      {className: classNames('popup', isNative && 'native', this.props.className)},
      this.props.children,
    )
  },
})
