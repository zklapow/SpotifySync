package main

import (
	"github.com/op/go-libspotify/spotify"
	"path"
	"os"
	"io/ioutil"
	"sync"
	"time"
	"os/signal"
	"github.com/gordonklaus/portaudio"
)

type Client struct {
	Conf      *Config
	Session   *spotify.Session
	portaudio *portAudio
}

func newClient(conf *Config) *Client {
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

	var wg sync.WaitGroup
	go func() {
		<-session.LoggedInUpdates()
		wg.Done()
	}()

	wg.Add(1)
	creds := spotify.Credentials{Username: conf.Username, Password: conf.Password}
	if err = session.Login(creds, false); err != nil {
		logger.Fatalf("Failed to log in to spotify: %v", err)
	}

	wg.Wait()

	return &Client{Conf: conf, Session: session, portaudio: pa}
}

func (client *Client) Run() {
	portaudio.Initialize()
	go client.portaudio.player()
	defer portaudio.Terminate()

	exit := make(chan bool)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	go func() {
		for _ = range signals {
			select {
			case exit <- true:
			default:
			}
		}
	}()

	go func() {

	}()

	user, err := client.Session.CurrentUser()
	if err != nil {
		logger.Fatalf("Could not get current user: %v", err)
	}
	user.Wait()

	list := client.Session.TracksToplist(spotify.ToplistRegionEverywhere)
	list.Wait()

	track := list.Track(0)

	logger.Infof("Got top list: %v", track)
	err = client.Session.Player().Load(track)
	if err != nil {
		logger.Fatalf("Error loading track: %v", err)
	}

	client.Session.Player().Play()

	client.handleSpotifyEvents(exit)
}

func (client *Client) handleSpotifyEvents(exit chan bool) {
	session := client.Session
	exitAttempts := 0
	running := true
	for running {
		logger.Debug("waiting for connection state change", session.ConnectionState())

		select {
		case err := <-session.LoggedInUpdates():
			logger.Debug("!! login updated", err)
		case <-session.LoggedOutUpdates():
			logger.Debug("!! logout updated")
			running = false
			break
		case err := <-session.ConnectionErrorUpdates():
			logger.Debug("!! connection error", err.Error())
		case msg := <-session.MessagesToUser():
			logger.Info("!! message to user", msg)
		case message := <-session.LogMessages():
			logger.Debug("!! log message", message.String())
		case _ = <-session.CredentialsBlobUpdates():
			logger.Debug("!! blob updated")
		case <-session.ConnectionStateUpdates():
			logger.Debug("!! connstate", session.ConnectionState())
		case <-exit:
			logger.Debug("!! exiting")
			if exitAttempts >= 3 {
				os.Exit(42)
			}
			exitAttempts++
			session.Logout()
		case <-time.After(5 * time.Second):
			println("state change timeout")
		}
	}

	session.Close()
	os.Exit(32)
}
