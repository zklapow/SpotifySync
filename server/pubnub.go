package main

import (
	"github.com/pubnub/go/messaging"
	"github.com/zklapow/SpotifySync/lib"
	"time"
)

type PubnubPublisher struct {
	conf *Config
	pubnub *messaging.Pubnub
	timeSyncer *lib.TimeSyncer
}

func newPubnubPublisher(conf *Config, pubnub *messaging.Pubnub) *PubnubPublisher {
	return &PubnubPublisher{conf: conf, pubnub: pubnub, timeSyncer: lib.StartTimeSync(pubnub)}
}

func (p *PubnubPublisher) Play(channel, trackUri string) {
	logger.Debugf("Publishing %v to channel %v", trackUri, channel)

	cmd := map[string]interface{}{
		"cmd":   lib.CommandTypePlay,
		"track": trackUri,
		"execAt": p.timeSyncer.SyncedTime().Add(time.Duration(p.conf.LatencyDelayMs) * time.Millisecond).UnixNano(),
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
