package main

type Events struct {
	play chan string
}

func newEvents() *Events {
	return &Events{
		play: make(chan string),
	}
}

func (events *Events) Play(track string) {
	events.play <- track
}

func (events *Events) PlayEvents() chan string {
	return events.play
}
