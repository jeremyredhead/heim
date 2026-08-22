import _ from 'lodash'
import Reflux from 'reflux'

const actions = Reflux.createActions([
  'sendMessage',
  'loadMoreLogs',
  'refreshUsers',
  'setNick',
  'tryRoomPasscode',
  'setup',
  'connect',
  'joinRoom',
  'embedMessage',
])
_.extend(module.exports, actions)

// sync so that we initialize room name / storage in the load tick
actions.setup.sync = true

// sync so that chatEntry can pass its state off to tentativeNick immediately after calling setNick
actions.setNick.sync = true

// sync so that embed components can react quickly to events
actions.embedMessage.sync = true
