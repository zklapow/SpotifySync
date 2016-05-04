package main

import (
	"github.com/pubnub/go/messaging"
	"encoding/json"
)

const (
	CommandTypeAdd = "add"
)

type PubNubEventDispatcher struct {
	events *Events
	pubnub *messaging.Pubnub
}

func newPubNubEventDispatcher(events *Events, conf *Config) *PubNubEventDispatcher {
	pubnub := messaging.NewPubnub(conf.PublishKey, conf.SubscribeKey, conf.SecretKey, "", false, "")

	return &PubNubEventDispatcher{events: events, pubnub: pubnub}
}

func (dispatch *PubNubEventDispatcher) Run() {
	cbChan := make(chan []byte)
	errChan := make(chan []byte)
	go dispatch.pubnub.Subscribe("test", "", cbChan, false, errChan)
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

func (dispatch *PubNubEventDispatcher) handleCommand(command map[string]interface{}) {
	cmdType, ok := command["cmd"]
	if !ok {
		logger.Errorf("Command from pubnub was malformed, no command type specified!")
		return
	}

	switch cmdType {
	case CommandTypeAdd:
		track, ok := command["track"]
		if !ok {
			logger.Errorf("Expected field 'track' was not present in add command!")
			return
		}
		logger.Debugf("Enqueuing link %v", track)

		dispatch.events.EnqueueLink(track.(string))
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

