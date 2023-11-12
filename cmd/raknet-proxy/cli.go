package main

import (
	"fmt"

	_cli "github.com/urfave/cli/v2"

	"github.com/percygrunwald/raknet-proxy/lib/cli"
)

var (
	flagValueLogLevel       string
	flagValueLogFormat      string
	flagValueServerHostname string
	flagValueServerPort     int
	flagValueListenPort     int
)

var cliFlags = []_cli.Flag{
	&_cli.IntFlag{
		Name:        "listen-port",
		Usage:       "Port on which to listen for RakNet packets from clients",
		Required:    true,
		Action:      cli.ValidatePort,
		Destination: &flagValueListenPort,
	},
	&_cli.StringFlag{
		Name:        "log-format",
		Usage:       fmt.Sprintf("Format in which to output logs. Valid options: %v", cli.LogFormats),
		Value:       cli.DefaultLogFormat.Text,
		Action:      cli.ValidateLogFormat,
		Destination: &flagValueLogFormat,
	},
	&_cli.StringFlag{
		Name:        "log-level",
		Usage:       fmt.Sprintf("Set the log level. Valid options: %v", cli.LogLevels),
		Value:       cli.DefaultLogLevel.Text,
		Action:      cli.ValidateLogLevel,
		Destination: &flagValueLogLevel,
	},
	&_cli.StringFlag{
		Name:        "server-hostname",
		Usage:       "Hostname/IP of upstream server",
		Required:    true,
		Destination: &flagValueServerHostname,
	},
	&_cli.IntFlag{
		Name:        "server-port",
		Usage:       "Upstream server RakNet port",
		Required:    true,
		Action:      cli.ValidatePort,
		Destination: &flagValueServerPort,
	},
}
