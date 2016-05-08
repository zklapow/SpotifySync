package main

import (
	"encoding/json"
	"github.com/pubnub/go/messaging"
	"github.com/zklapow/SpotifySync/lib"
)

type PubNubEventDispatcher struct {
	conf   *Config
	events *Events
	pubnub *messaging.Pubnub
}

func newPubNubEventDispatcher(events *Events, pubnub *messaging.Pubnub, conf *Config) *PubNubEventDispatcher {
	return &PubNubEventDispatcher{conf: conf, events: events, pubnub: pubnub}
}

func (dispatch *PubNubEventDispatcher) Run() {
	logger.Infof("Connecting to channel: %v", dispatch.conf.Channel)
	cbChan := make(chan []byte)
	errChan := make(chan []byte)
	go dispatch.pubnub.Subscribe(dispatch.conf.Channel, "", cbChan, false, errChan)
	go dispatch.handleSubscribeResult(cbChan, errChan)
}

func (dispatch *PubNubEventDispatcher) handleSubscribeResult(successChannel, errorChannel chan []byte) {
	for {
		select {
		case success, ok := <-successChannel:
			if !ok {
				break
			}
			if string(success) != "[]" {
				logger.Debugf("Successfully subscribed from pubnub channel: %v", string(success))
				dispatch.handleCommand(decodeCommand(success))
			}
		case failure, ok := <-errorChannel:
			if !ok {
				break
			}
			if string(failure) != "[]" {
				logger.Errorf("Failed to subscribe to pubnub channel: %v", string(failure))
			}
		}
	}
}

func (dispatch *PubNubEventDispatcher) handlePublishResult(successChannel, errorChannel chan []byte) {
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

func (dispatch *PubNubEventDispatcher) handleCommand(command map[string]interface{}) {
	cmdType, ok := command["cmd"]
	if !ok {
		logger.Errorf("Command from pubnub was malformed, no command type specified!")
		return
	}

	switch cmdType {
	case lib.CommandTypePlay:
		track, ok := command["track"]
		if !ok {
			logger.Errorf("Expected field 'track' was not present in add command!")
			return
		}
		logger.Debugf("Enqueuing link %v", track)

		dispatch.events.Play(track.(string))
	}
}

func decodeCommand(data []byte) map[string]interface{} {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		logger.Errorf("Failed to parse command from pubnub: %v", err)
	}

	return decodeCommandArray(arr)
}

func decodeCommandArray(in interface{}) map[string]interface{} {
	switch vv := in.(type) {
	case []interface{}:
		for _, u := range vv {
			return decodeCommandArray(u)
		}
	case map[string]interface{}:
		return vv
	default:
	}

	return nil
}
