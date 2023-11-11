package main

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type logLevel struct {
	text  string
	level log.Level
}

func (l logLevel) String() string {
	return l.text
}

type logFormat struct {
	text      string
	formatter log.Formatter
}

func (l logFormat) String() string {
	return l.text
}

const (
	portMax int = 65535
)

var (
	logLevelTrace = logLevel{text: "trace", level: log.TraceLevel}
	logLevelDebug = logLevel{text: "debug", level: log.DebugLevel}
	logLevelInfo  = logLevel{text: "info", level: log.InfoLevel}
	logLevelWarn  = logLevel{text: "warn", level: log.WarnLevel}
	logLevelError = logLevel{text: "error", level: log.ErrorLevel}
	logLevels     = []logLevel{logLevelTrace, logLevelDebug, logLevelInfo,
		logLevelWarn, logLevelError}
	defaultLogLevel = logLevelInfo
)

var (
	logFormatText    = logFormat{text: "text", formatter: &log.TextFormatter{}}
	logFormatJSON    = logFormat{text: "json", formatter: &log.JSONFormatter{}}
	logFormats       = []logFormat{logFormatText, logFormatJSON}
	defaultLogFormat = logFormatJSON
)

var (
	flagValueLogLevel   string
	flagValueLogFormat  string
	flagValueListenPort int
)

var cliFlags = []cli.Flag{
	&cli.IntFlag{
		Name:        "listen-port",
		Usage:       "Port on which to listen for RakNet packets from clients",
		Required:    true,
		Action:      validatePort,
		Destination: &flagValueListenPort,
	},
	&cli.StringFlag{
		Name:        "log-format",
		Usage:       fmt.Sprintf("Format in which to output logs. Valid options: %v", logFormats),
		Value:       defaultLogFormat.text,
		Action:      validateLogFormat,
		Destination: &flagValueLogFormat,
	},
	&cli.StringFlag{
		Name:        "log-level",
		Usage:       fmt.Sprintf("Set the log level. Valid options: %v", logLevels),
		Value:       defaultLogLevel.text,
		Action:      validateLogLevel,
		Destination: &flagValueLogLevel,
	},
}

func validatePort(ctx *cli.Context, v int) error {
	if v < 0 || v > portMax {
		return fmt.Errorf(`Invalid port value: %d. Valid port range: [0-%d]`, v, portMax)
	}
	return nil
}

func validateLogLevel(ctx *cli.Context, v string) error {
	return newValidateStringOption[logLevel](logLevels)(ctx, v)
}

func validateLogFormat(ctx *cli.Context, v string) error {
	return newValidateStringOption[logFormat](logFormats)(ctx, v)
}

// newValidateStringOption returns a function that can be used as a validator
// for a urfave/cli string flag. The returned function checks that the option
// entered by the user matches one of the items in a list of structs, based on
// the string representation of each item.
func newValidateStringOption[T fmt.Stringer](list []T) func(ctx *cli.Context, v string) error {
	return func(ctx *cli.Context, v string) error {
		for _, item := range list {
			if item.String() == v {
				return nil
			}
		}
		return fmt.Errorf(`Invalid option for %v "%s". Valid options: %v`, reflect.TypeOf(list).Elem().Name(), v, list)
	}
}

func getLogLevel(s string) logLevel {
	return newGetOption[logLevel](logLevels, defaultLogLevel)(s)
}

func getLogFormat(s string) logFormat {
	return newGetOption[logFormat](logFormats, defaultLogFormat)(s)
}

// newGetOption returns a function that can be used to get a struct in a list
// of options from its string representation. It is intended for going from a
// string input from a command line flag into a struct. E.g. if a user passes
// a log level option of "debug", this function can be used to generate a
// function that gets the logLevelDebug struct out of the list of all logLevel
// structs.
func newGetOption[T fmt.Stringer](list []T, defaultOption T) func(s string) T {
	return func(s string) T {
		for _, item := range list {
			if item.String() == s {
				return item
			}
		}
		return defaultOption
	}
}
