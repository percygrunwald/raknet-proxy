package cli

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

type LogLevel struct {
	Text  string
	Level log.Level
}

func (l LogLevel) String() string {
	return l.Text
}

type LogFormat struct {
	Text      string
	Formatter log.Formatter
}

func (l LogFormat) String() string {
	return l.Text
}

const (
	portMax int = 65535
)

var (
	LogLevelTrace = LogLevel{Text: "trace", Level: log.TraceLevel}
	LogLevelDebug = LogLevel{Text: "debug", Level: log.DebugLevel}
	LogLevelInfo  = LogLevel{Text: "info", Level: log.InfoLevel}
	LogLevelWarn  = LogLevel{Text: "warn", Level: log.WarnLevel}
	LogLevelError = LogLevel{Text: "error", Level: log.ErrorLevel}
	LogLevels     = []LogLevel{LogLevelTrace, LogLevelDebug, LogLevelInfo,
		LogLevelWarn, LogLevelError}
	DefaultLogLevel = LogLevelInfo
)

var (
	LogFormatText    = LogFormat{Text: "text", Formatter: &log.TextFormatter{}}
	LogFormatJSON    = LogFormat{Text: "json", Formatter: &log.JSONFormatter{}}
	LogFormats       = []LogFormat{LogFormatText, LogFormatJSON}
	DefaultLogFormat = LogFormatJSON
)

func ValidatePort(ctx *cli.Context, v int) error {
	if v < 0 || v > portMax {
		return fmt.Errorf(`Invalid port value: %d. Valid port range: [0-%d]`, v, portMax)
	}
	return nil
}

func ValidateLogLevel(ctx *cli.Context, v string) error {
	return newValidateStringOption[LogLevel](LogLevels)(ctx, v)
}

func ValidateLogFormat(ctx *cli.Context, v string) error {
	return newValidateStringOption[LogFormat](LogFormats)(ctx, v)
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

func GetLogLevel(s string) LogLevel {
	return newGetOption[LogLevel](LogLevels, DefaultLogLevel)(s)
}

func GetLogFormat(s string) LogFormat {
	return newGetOption[LogFormat](LogFormats, DefaultLogFormat)(s)
}

// newGetOption returns a function that can be used to get a struct in a list
// of options from its string representation. It is intended for going from a
// string input from a command line flag into a struct. E.g. if a user passes
// a log level option of "debug", this function can be used to generate a
// function that gets the LogLevelDebug struct out of the list of all LogLevel
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
