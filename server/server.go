package main

import (
	"github.com/nlopes/slack"
	"github.com/op/go-libspotify/spotify"
	"github.com/zklapow/SpotifySync/lib"
	"io/ioutil"
	"regexp"
	"strings"
)

var REGEX *regexp.Regexp

func init() {
	var err error
	REGEX, err = regexp.Compile(`spotify:track:[\w\d]+`)

	if err != nil {
		logger.Fatalf("Failure constructing regex %v", err)
	}
}

type Server struct {
	conf      *Config
	publisher *PubnubPublisher
	queue     *lib.PlayQueue
	session   *spotify.Session
	state     *PlayState
}

func newServer(conf *Config) *Server {
	return &Server{
		conf:      conf,
		publisher: newPubnubPublisher(conf),
		queue:     lib.NewPlayQueue(),
		session:   createSpotifySession(conf),
	}
}

func (server *Server) Run() {
	api := slack.New(server.conf.SlackToken)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	user, err := server.session.CurrentUser()
	if err != nil {
		logger.Fatalf("Could not get current user: %v", err)
	}
	user.Wait()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
			// Ignore hello

			case *slack.ConnectedEvent:
				logger.Debugf("Infos: %v", ev.Info)
				logger.Debugf("Connection counter: %v", ev.ConnectionCount)
				// Replace #general with your Channel ID
				rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "#general"))

			case *slack.MessageEvent:
				logger.Debugf("Message: %v\n", ev)

				channel, err := api.GetChannelInfo(ev.Channel)
				if err != nil {
					logger.Errorf("Failed to get channel info: %v", err)
					continue
				}

				if channel.Name != server.conf.Channel {
					logger.Debugf("Skipping message not in correct channel")
					continue
				}

				match := REGEX.FindString(ev.Text)
				if match != "" {
					server.queueOrPlay(match)
				} else if strings.ToLower(ev.Text) == "skip" {
					server.playNext()
				}

			case *slack.PresenceChangeEvent:
				logger.Debugf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				logger.Debugf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				logger.Errorf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				logger.Fatalf("Invalid credentials")

			default:
				logger.Warningf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

func (server *Server) queueOrPlay(linkString string) {
	logger.Debugf("Queueing or playing %v", linkString)
	if server.state == nil {
		server.play(linkString)
	} else {
		server.queue.Append(linkString)
	}
}

func (server *Server) play(trackString string) {
	logger.Debugf("Playing %v", trackString)
	server.state = newPlayState(trackString, server.session)

	go server.handleStateTimeout(*server.state)

	server.publisher.Play(server.conf.Channel, trackString)
}

func (server *Server) handleStateTimeout(state PlayState) {
	<-state.End()
	logger.Debugf("Reached end of track %v", state.TrackLink)
	server.state = nil
}

func (server *Server) playNext() {
	if !server.queue.IsEmpty() {
		trackString := server.queue.Pop()
		if trackString != "" {
			server.play(trackString)
		}
	}
}

func createSpotifySession(conf *Config) *spotify.Session {
	appKey, err := ioutil.ReadFile(conf.AppKeyPath)
	if err != nil {
		logger.Fatalf("Failed to load appkey: %v", err)
	}

	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   appKey,
		ApplicationName:  "spotifysync",
		CacheLocation:    "/tmp/server",
		SettingsLocation: "/tmp/server",
	})

	if err != nil {
		logger.Fatalf("Error establishing spotify session: %v", err)
	}

	loginChan := make(chan bool)
	go func() {
		<-session.LoggedInUpdates()
		loginChan <- true
	}()

	creds := spotify.Credentials{Username: conf.Username, Password: conf.Password}
	if err = session.Login(creds, false); err != nil {
		logger.Fatalf("Failed to log in to spotify: %v", err)
	}

	<-loginChan

	return session
}

func handlePublishResult(successChannel, errorChannel chan []byte) {
	for {
		select {
		case success, ok := <-successChannel:
			if !ok {
				break
			}
			if string(success) != "[]" {
				logger.Debugf("Successfully published to pubnub channel: %v", string(success))
			}
		case failure, ok := <-errorChannel:
			if !ok {
				break
			}
			if string(failure) != "[]" {
				logger.Errorf("Failed to publish to pubnub channel: %v", string(failure))
			}
		}
	}
}
