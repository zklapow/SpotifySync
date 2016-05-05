package main

import (
	"github.com/op/go-libspotify/spotify"
	"time"
)

type PlayState struct {
	TrackLink string
	PlayTimer *time.Timer
}

func newPlayState(linkString string, session *spotify.Session) *PlayState {
	link, err := session.ParseLink(linkString)
	if err != nil {
		logger.Errorf("Failed to create playstate for %v: %v", linkString, err)
		return nil
	}

	logger.Debugf("Parsed link")

	track, err := link.Track()
	if err != nil {
		logger.Errorf("Failed to get track for link %v: %v", linkString, err)
		return nil
	}
	track.Wait()

	logger.Debugf("loaded track")

	return &PlayState{TrackLink: linkString, PlayTimer: time.NewTimer(track.Duration() + time.Second)}
}

func (state *PlayState) Skip() {
	state.PlayTimer.Stop()
}

func (state *PlayState) End() <-chan time.Time {
	return state.PlayTimer.C
}
