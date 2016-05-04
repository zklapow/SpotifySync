package main

import "github.com/op/go-libspotify/spotify"

type PlayQueue struct {
	elements []*spotify.Track
}


func newPlayQueue() *PlayQueue {
	return &PlayQueue{elements: make([]*spotify.Track, 0, 0)}
}

func (queue *PlayQueue) Append(track *spotify.Track) *spotify.Track {
	queue.elements = append(queue.elements, track)

	return track
}

func (queue *PlayQueue) Pop() *spotify.Track {
	if len(queue.elements) == 0 {
		return nil
	}

	track := queue.elements[0]
	queue.elements = queue.elements[1:len(queue.elements)]
	return track
}

func (queue *PlayQueue) IsEmpty() bool {
	return len(queue.elements) == 0
}