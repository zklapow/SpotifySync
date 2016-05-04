package main

import (
	"github.com/op/go-libspotify/spotify"
	"path"
	"os"
	"io/ioutil"
	"time"
	"github.com/gordonklaus/portaudio"
)

type SpotifyPlayer struct {
	Conf      *Config
	Session   *spotify.Session
	portaudio *portAudio
	exit 	  chan bool
}

func newSpotifyPlayer(conf *Config) *SpotifyPlayer {
	prog := path.Base(os.Args[0])
	appKey, err := ioutil.ReadFile(conf.AppKeyPath)
	if err != nil {
		logger.Fatalf("Failed to load appkey: %v", err)
	}

	pa := newPortAudio()

	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey: appKey,
		ApplicationName: prog,
		CacheLocation: "tmp",
		SettingsLocation: "tmp",
		AudioConsumer: pa,
	})

	if err != nil {
		logger.Fatalf("Error establishing spotify session: %v", err)
	}

	loginChan := make(chan bool, 1)
	go func() {
		<-session.LoggedInUpdates()
		loginChan <- true
	}()

	creds := spotify.Credentials{Username: conf.Username, Password: conf.Password}
	if err = session.Login(creds, false); err != nil {
		logger.Fatalf("Failed to log in to spotify: %v", err)
	}

	<-loginChan

	return &SpotifyPlayer{Conf: conf, Session: session, portaudio: pa, exit: make(chan bool)}
}

func (player *SpotifyPlayer) Run() {
	portaudio.Initialize()
	go player.portaudio.player()
	defer portaudio.Terminate()

	user, err := player.Session.CurrentUser()
	if err != nil {
		logger.Fatalf("Could not get current user: %v", err)
	}
	user.Wait()

	list := player.Session.TracksToplist(spotify.ToplistRegionEverywhere)
	list.Wait()

	track := list.Track(0)

	logger.Infof("Got top list: %v", track)
	err = player.Session.Player().Load(track)
	if err != nil {
		logger.Fatalf("Error loading track: %v", err)
	}

	player.Session.Player().Play()

	player.handleSpotifyEvents()
}

func (player *SpotifyPlayer) Close() {
	player.exit <- true
}

func (player *SpotifyPlayer) handleSpotifyEvents() {
	session := player.Session
	exitAttempts := 0
	running := true
	for running {
		logger.Debug("waiting for connection state change", session.ConnectionState())

		select {
		case err := <-session.LoggedInUpdates():
			logger.Debugf("login updated: %v", err)
		case <-session.LoggedOutUpdates():
			logger.Debug("logged out")
			running = false
			break
		case err := <-session.ConnectionErrorUpdates():
			logger.Errorf("connection error: %v", err.Error())
		case msg := <-session.MessagesToUser():
			logger.Infof("message to user: %v", msg)
		case message := <-session.LogMessages():
			logger.Debugf("log message: %v", message.String())
		case _ = <-session.CredentialsBlobUpdates():
			logger.Debug("blob updated")
		case <-session.ConnectionStateUpdates():
			logger.Debugf("connstate: %v", session.ConnectionState())
		case <-player.exit:
			logger.Debug("exiting")
			if exitAttempts >= 3 {
				os.Exit(42)
			}
			exitAttempts++
			session.Logout()
		case <-time.After(5 * time.Second):
			logger.Debug("state change timeout")
		}
	}

	session.Close()
	os.Exit(32)
}
