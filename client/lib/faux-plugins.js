var React = require('react/addons')
var Reflux = require('reflux')
var Immutable = require('immutable')


// allow var redeclaration for import dupes
// jshint -W004

module.exports = function(roomName) {
  if (roomName == 'thedrawingroom' || roomName == 'lovenest' || roomName == 'has') {
    Heim.hook('page-bottom', function() {
      return (
        <style key="drawingroom-style" dangerouslySetInnerHTML={{__html:`
          .chat-pane.timestamps-visible {
            background: #333;
          }

          .main-pane .room .name,
          .info-pane .thread-list .thread .info .title {
            color: #222;
          }

          .chat-pane time {
            opacity: .5;
          }

          .main-pane .room .state,
          .nick {
            background: #e8e8e8 !important;
          }

          .message-emote {
            background: #f3f3f3 !important;
          }

          .mention-nick {
            color: #000 !important;
            font-weight: bold;
          }

          a {
            color: #444;
            text-decoration: none;
            font-weight: bold;
          }
        `}} />
      )
    })
  }

  if (roomName == 'space') {
    var Embed = require('./ui/embed')

    Heim.hook('main-sidebar', function() {
      // jshint camelcase: false
      return (
        <div key="norman" className="norman">
          <p>norman</p>
          <Embed kind="imgur" imgur_id="UKbitCO" />
        </div>
      )
    })

    Heim.hook('page-bottom', function() {
      return (
        <style key="norman-style" dangerouslySetInnerHTML={{__html:`
          .norman {
            text-align: right;
            opacity: .5;
          }

          .norman, .norman .embed {
            transition: transform .15s ease;
          }

          .norman:hover {
            opacity: 1;
          }

          .norman:hover .embed {
            transform: translate(-50%, 50%) scale(2);
          }

          .norman p {
            margin: 0;
            font-size: 12px;
          }

          .norman .embed {
            width: 0;
            height: 100px;
            border: none;
          }
        `}} />
      )
    })
  }

  if (roomName == 'music' || roomName == 'youtube') {
    var Embed = require('./ui/embed')
    var MessageText = require('./ui/message-text')

    var clientTimeOffset = 0
    Heim.socket.store.listen(function(ev) {
      if (ev.status == 'receive' && ev.body.type == 'ping-event') {
        clientTimeOffset = Date.now() / 1000 - ev.body.data.time
      }
    })

    var TVActions = Reflux.createActions([
      'changeVideo',
      'changeNotice',
    ])

    Heim.ui.createCustomPane('youtube-tv', {readOnly: true})

    var TVStore = Reflux.createStore({
      listenables: [
        TVActions,
        {chatChange: Heim.chat.store},
      ],

      init: function() {
        this.state = Immutable.fromJS({
          video: {
            time: 0,
            messageId: null,
            youtubeId: null,
            title: '',
          },
          notice: {
            time: 0,
            content: '',
          },
        })
      },

      getInitialState: function() {
        return this.state
      },

      changeVideo: function(video) {
        this.state = this.state.set('video', Immutable.fromJS(video))
        this.trigger(this.state)
      },

      changeNotice: function(notice) {
        this.state = this.state.set('notice', Immutable.fromJS(notice))
        this.trigger(this.state)
      },
    })

    var SyncedEmbed = React.createClass({
      displayName: 'SyncedEmbed',

      shouldComponentUpdate: function(nextProps) {
        return nextProps.youtubeId != this.props.youtubeId
      },

      render: function() {
        // jshint camelcase: false
        return (
          <Embed
            className={this.props.className}
            kind="youtube"
            autoplay="1"
            start={Math.max(0, Math.floor(Date.now() / 1000 - this.props.startedAt - clientTimeOffset))}
            youtube_id={this.props.youtubeId}
          />
        )
      }
    })

    var YouTubePane = React.createClass({
      displayName: 'YouTubePane',

      mixins: [
        Reflux.connect(TVStore, 'tv'),
        React.addons.PureRenderMixin,
      ],

      render: function() {
        // jshint camelcase: false
        return (
          <div className="chat-pane-container youtube-pane">
            <div className="top-bar">
              <MessageText className="title" content={':notes: :tv: :notes: ' + this.state.tv.getIn(['video', 'title'])} />
            </div>
            <div className="aspect-wrapper">
              <SyncedEmbed
                className="youtube-tv"
                youtubeId={this.state.tv.getIn(['video', 'youtubeId'])}
                startedAt={this.state.tv.getIn(['video', 'time'])}
              />
            </div>
            <MessageText className="notice" content={this.state.tv.getIn(['notice', 'content'])} />
          </div>
        )
      }
    })

    Heim.hook('thread-panes', function() {
      return <YouTubePane key="youtube-tv" />
    })

    Heim.chat.messagesChanged.listen(function(ids, state) {
      var candidates = Immutable.Seq(ids)
        .map(messageId => {
          var msg = state.messages.get(messageId)
          var valid = messageId != '__root' && msg.get('content')
          return valid && msg
        })
        .filter(Boolean)

      var playRe = /!play [^?]*\?v=([-\w]+)/
      var video = candidates
        .map(msg => {
          var match = msg.get('content').match(playRe)
          return match && {
            time: msg.get('time'),
            messageId: msg.get('id'),
            youtubeId: match[1],
            title: msg.get('content'),
          }
        })
        .filter(Boolean)
        .sortBy(video => video.time)
        .last()

      if (video && video.time > TVStore.state.getIn(['video', 'time'])) {
        TVActions.changeVideo(video)
      }

      var noticeRe = /^!notice ([^]*)$/
      var notice = candidates
        .map(msg => {
          var match = msg.get('content').match(noticeRe)
          return match && {
            time: msg.get('time'),
            content: match[1],
          }
        })
        .filter(Boolean)
        .sortBy(notice => notice.time)
        .last()

      if (notice && notice.time > TVStore.state.getIn(['notice', 'time'])) {
        TVActions.changeNotice(notice)
      }
    })

    Heim.hook('page-bottom', function() {
      return (
        <style key="youtubetv-style" dangerouslySetInnerHTML={{__html:`
          .youtube-pane {
            z-index: 9;
          }

          .youtube-pane .title {
            width: 0;
          }

          .youtube-pane .aspect-wrapper {
            flex-shrink: 0;
            position: relative;
            width: 100%;
            box-shadow: 0 0 12px rgba(0, 0, 0, .25);
            z-index: 5;
          }

          .youtube-pane .aspect-wrapper:before {
            content: '';
            display: block;
            padding-top: 75%;
          }

          .youtube-pane .youtube-tv {
            position: absolute;
            top: 0;
            bottom: 0;
            left: 0;
            right: 0;
            width: 100%;
            height: 100%;
            border: none;
          }

          .youtube-pane .notice {
            background: white;
            padding: 10px;
            overflow: auto;
            white-space: pre-wrap;
            flex: 1;
          }
        `}} />
      )
    })
  }

  if (roomName == 'adventure' || roomName == 'chess') {
    Heim.hook('page-bottom', function() {
      return (
        <style key="adventure-style" dangerouslySetInnerHTML={{__html:`
          .messages-container, .messages-container input, .messages-container textarea {
            font-family: Droid Sans Mono, monospace;
          }
        `}} />
      )
    })

    Heim.chat.setRoomSettings({collapse: false})
  }
}
