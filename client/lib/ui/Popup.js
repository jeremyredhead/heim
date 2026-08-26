import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'
import classNames from 'classnames'

export default createReactClass({
  displayName: 'Popup',

  propTypes: {
    kind: PropTypes.string,
    className: PropTypes.string,
    onClick: PropTypes.func,
    onInnerClick: PropTypes.func,
    onDismiss: PropTypes.func,
    children: PropTypes.node,
  },

  componentDidMount() {
    const node = ReactDOM.findDOMNode(this)
    if (this.props.kind === 'native') {
      try {
        node.show()
      } catch (e) {
        node.setAttribute('open', '')
      }
    } else if (this.props.kind === 'native-modal') {
      try {
        node.showModal()
      } catch (e) {
        node.setAttribute('open', '')
      }
    }
    Heim.addEventListener(uidocument, Heim.isTouch ? 'touchstart' : 'click', this.onOutsideClick, false)
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
      {className: classNames('popup', isNative && 'native', this.props.className), tabIndex: -1, onClick: this.props.onClick},
      isNative ? <div className="popup-content" onClick={this.props.onInnerClick}>{this.props.children}</div> : this.props.children,
    )
  },
})
