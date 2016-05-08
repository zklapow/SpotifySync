package main

import (
	"github.com/pubnub/go/messaging"
	"github.com/zklapow/SpotifySync/lib"
)

type PubnubPublisher struct {
	pubnub *messaging.Pubnub
}

func newPubnubPublisher(conf *Config, pubnub *messaging.Pubnub) *PubnubPublisher {
	return &PubnubPublisher{pubnub: pubnub}
}

func (p *PubnubPublisher) Play(channel, trackUri string) {
	logger.Debugf("Publishing %v to channel %v", trackUri, channel)
	cmd := map[string]string{
		"cmd":   lib.CommandTypePlay,
		"track": trackUri,
	}

	go p.publish(channel, cmd)
}

func (p *PubnubPublisher) Skip(channel string) {
	cmd := map[string]string{
		"cmd": "skip",
	}

	go p.publish(channel, cmd)
}

func (p *PubnubPublisher) publish(channel string, cmd interface{}) {
	cbChan := make(chan []byte)
	errChan := make(chan []byte)

	go func() {
		success := <-cbChan
		logger.Infof("Got success from publish: %v", string(success))
	}()

	go func() {
		err := <-errChan
		logger.Infof("Got failure from publish: %v", string(err))
	}()

	p.pubnub.Publish(channel, cmd, cbChan, errChan)
}
