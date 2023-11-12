package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/sandertv/go-raknet"
	log "github.com/sirupsen/logrus"
	_cli "github.com/urfave/cli/v2"

	"github.com/percygrunwald/raknet-proxy/lib/cli"
)

func main() {
	app := &_cli.App{
		Name:    "raknet-test-client",
		Usage:   "Test client that tries to connect to a server",
		Flags:   cliFlags,
		Action:  runApp,
		Version: "v0.0.1",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApp(cCtx *_cli.Context) error {
	logLevel := cli.GetLogLevel(flagValueLogLevel)
	logFormat := cli.GetLogFormat(flagValueLogFormat)
	log.SetFormatter(logFormat.Formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel.Level)

	serverAddr := fmt.Sprintf("%s:%d", flagValueServerHostname, flagValueServerPort)
	log.Debugf("dialing %v", serverAddr)
	conn, err := raknet.Dial(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to dial %v: %w", serverAddr, err)
	}

	log.Debugf("connected to %v", serverAddr)
	defer conn.Close()

	return nil
}
