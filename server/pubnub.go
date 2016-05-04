package main

import "github.com/pubnub/go/messaging"

type PubnubPublisher struct  {
	pubnub *messaging.Pubnub
}

func newPubnubPublisher(conf *Config) *PubnubPublisher {
	pubnub := messaging.NewPubnub(conf.PublishKey, conf.SubscribeKey, conf.SecretKey, "", false, "")
	return &PubnubPublisher{pubnub: pubnub}
}

func (p *PubnubPublisher) AddTrack(channel, trackUri string) {
	logger.Debugf("Publishing to channel %v: %v", channel, trackUri)
	cbChan := make(chan []byte)
	errChan := make(chan []byte)

	go func() {
		success := <- cbChan
		logger.Infof("Got success from publish: %v", string(success))
	}()

	go func() {
		err := <- errChan
		logger.Infof("Got failure from publish: %v", string(err))
	}()

	cmd := map[string]string{
		"cmd": "add",
		"track": trackUri,
	}

	p.pubnub.Publish(channel, cmd, cbChan, errChan)
}
