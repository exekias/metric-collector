package logging

import glogging "github.com/op/go-logging"

var format = glogging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	glogging.SetFormatter(format)
}

// MustGetLogger returns a logger instance for the given module name
func MustGetLogger(module string) *glogging.Logger {
	return glogging.MustGetLogger(module)
}
