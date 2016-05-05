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
	cmd := map[string]string{
		"cmd": "add",
		"track": trackUri,
	}

	p.publish(channel, cmd)
}

func (p *PubnubPublisher) Skip(channel string) {
	cmd := map[string]string{
		"cmd": "skip",
	}

	p.publish(channel, cmd)
}

func (p *PubnubPublisher) publish(channel string, cmd interface{}) {
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

	p.pubnub.Publish(channel, cmd, cbChan, errChan)

}
