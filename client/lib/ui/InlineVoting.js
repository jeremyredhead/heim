import _ from 'lodash'
import React from 'react'
import twemoji from 'twemoji'

import actions from '../actions'
import Tree from '../Tree'
import FastButton from './FastButton'
import Message from './Message'
import MessageText from './MessageText'

export default React.createClass({
  displayName: 'InlineVoting',

  propTypes: {
    message: React.PropTypes.instanceOf(Message).isRequired,
    tree: React.PropTypes.instanceOf(Tree).isRequired,
    className: React.PropTypes.string,
    title: React.PropTypes.string,
    style: React.PropTypes.string
  },

  upvote(evt) {
    actions.sendMessage('+1', this.props.message.state.node.get('id'))
    if (evt) evt.stopPropagation();
  },

  downvote(evt) {
    actions.sendMessage('-1', this.props.message.state.node.get('id'))
    if (evt) evt.stopPropagation();
  },

  render() {
    let upvotes = 0, downvotes = 0

    this.props.message.state.node.get('children').map(id => {
      const content = this.props.tree.get(id).get('content')

      if (/\s*\+1\s*/.test(content)) upvotes++
      if (/\s*-1\s*/.test(content)) downvotes++
    })

    const result = (upvotes > downvotes) ? "approved" : (upvotes < downvotes) ? "rejected" : "neutral";

    return <span className={"inline-voting"}>
      <FastButton onClick={this.upvote} className='approve'>
        <MessageText content={':thumbsup:'} onlyEmoji /> {upvotes}
      </FastButton>
      <FastButton onClick={this.downvote} className='disapprove'>
        <MessageText content={':thumbsdown:'} onlyEmoji /> {downvotes}
      </FastButton>
      <span className={result}> {upvotes - downvotes}</span>
    </span>
  }

})
