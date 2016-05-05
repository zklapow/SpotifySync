package main

import (
	"github.com/nlopes/slack"
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
	conf *Config
	publisher *PubnubPublisher
}

func newServer(conf *Config) *Server {
	return &Server{conf: conf, publisher: newPubnubPublisher(conf)}
}

func (server *Server) Run() {
	api := slack.New(server.conf.SlackToken)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

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

				match := REGEX.FindString(ev.Text)
				if match != "" {
					server.publisher.AddTrack(channel.Name, match)
				} else if strings.ToLower(ev.Text) == "skip" {
					server.publisher.Skip(channel.Name)
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
