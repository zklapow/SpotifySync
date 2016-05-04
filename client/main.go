package main

import (
	"github.com/op/go-libspotify/spotify"
	"os"
	"github.com/lunixbochs/go-keychain"
	"github.com/BurntSushi/toml"
	"flag"
	"io/ioutil"
	"github.com/bgentry/speakeasy"
	"strings"
	"github.com/op/go-logging"
	"os/user"
	"path"
	"os/signal"
)

type Config struct {
	Coordinator string
	Username    string
	Password    string
	AppKeyPath  string
}

var logger = logging.MustGetLogger("spotifysync")

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

	configPath := flag.String("config", path.Join(usr.HomeDir, ".spotifysync.toml"), "the path to the spotify sync config file")

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
	logger.Infof("Starting sync client with config: %v", conf)

	getPassword(&conf)

	client := newSpotifyPlayer(&conf)
	client.Run()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	for _ = range signals {
		select {
		case client.exit <- true:
		default:
		}
	}
}

func getPassword(conf *Config) {
	password, err := keychain.Find("spotifysync", conf.Username)
	password = strings.TrimSpace(password)
	if password == "" || err != nil {
		logger.Debugf("could not find password in keychain")
		password, err = speakeasy.Ask("Enter spotify password (this will be stored in your keychain): ")
		if err != nil {
			logger.Fatalf("Failed to get password: %v", err)
			os.Exit(-1)
		}

		keychain.Add("spotifysync", conf.Username, password)
	}
	conf.Password = password
}
