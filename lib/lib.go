package lib

import "github.com/op/go-logging"

var logger = logging.MustGetLogger("spotifysync")

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
