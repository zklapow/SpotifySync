package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/op/go-libspotify/spotify"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"time"
	"math"
	"github.com/zklapow/SpotifySync/lib"
	"github.com/pubnub/go/messaging"
)

type Config struct {
	SlackToken   string
	PublishKey   string
	SubscribeKey string
	SecretKey    string
	AppKeyPath   string
	Username     string
	Password     string
	Channel      string
}

var logger = logging.MustGetLogger("syncserver")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func main() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)

	usr, err := user.Current()
	if err != nil {
		logger.Fatalf("Could not get current user: %v", err)
	}

	configPath := flag.String("config", path.Join(usr.HomeDir, ".spotifysyncserver.toml"), "the path to the spotify sync config file")

	flag.Parse()

	var conf Config
	logger.Debugf("Looking for config file at %v", *configPath)
	_, err = os.Stat(*configPath)
	if err == nil {
		fp, err := os.Open(*configPath)
		if err != nil {
			logger.Errorf("could not open config file %v", err)
			os.Exit(-1)
		}

		confData, err := ioutil.ReadAll(fp)
		if err := toml.Unmarshal(confData, &conf); err != nil {
			logger.Errorf("Error reading config: %v", err)
			os.Exit(-1)
		}
	} else {
		logger.Debug("No config file found, using empty config")
		conf = Config{}
	}

	logger.Debugf("lib spotify version: %v", spotify.BuildId())
	logger.Infof("Starting sync server with config: %v", conf)

	pubnub := messaging.NewPubnub(conf.PublishKey, conf.SubscribeKey, conf.SecretKey, "", false, "")

	timeSyncer := lib.StartTimeSync(pubnub)
	timeSyncer.AwaitSynced()

	now := time.Now()
	clientTime := timeSyncer.SyncedTime()
	logger.Infof("Time sync started (%v, %v, %v)", now, clientTime, math.Abs(float64(now.Unix()) - float64(clientTime.Unix())))

	s := newServer(&conf, newPubnubPublisher(&conf, pubnub), timeSyncer)
	s.Run()
}
