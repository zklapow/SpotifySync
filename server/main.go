package main

import (
	"os"
	"os/user"
	"flag"
	"path"
	"io/ioutil"
	"github.com/BurntSushi/toml"
	"github.com/op/go-libspotify/spotify"
	"github.com/op/go-logging"
)

type Config struct {
	SlackToken   string
	PublishKey   string
	SubscribeKey string
	SecretKey    string
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

	logger.Debug("lib spotify version: %v", spotify.BuildId());
	logger.Infof("Starting sync server with config: %v", conf)

	s := newServer(&conf)
	s.Run()
}