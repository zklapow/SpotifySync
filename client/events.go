package main

import "github.com/op/go-libspotify/spotify"

type Events struct {
	queue     chan *spotify.Track
	linkqueue chan string
	advance   chan bool
	skip      chan bool
}

func newEvents() *Events {
	return &Events{
		queue: make(chan *spotify.Track),
		linkqueue: make(chan string),
		advance: make(chan bool),
		skip: make(chan bool),
	}
}

func (events *Events) TriggerAdvance() {
	events.advance <- true
}

func (events *Events) AdvanceEvents() chan bool {
	return events.advance
}

func (events *Events) Enqueue(track *spotify.Track) {
	events.queue <- track
}

func (events *Events) EnqueueEvents() chan *spotify.Track {
	return events.queue
}

func (events *Events) EnqueueLink(track string) {
	events.linkqueue <- track
}

func (events *Events) LinkQueueEvents() chan string {
	return events.linkqueue
}

func (events *Events) Skip() {
	events.skip <- true
}

func (events *Events) SkipEvents() chan bool {
	return events.skip
}
