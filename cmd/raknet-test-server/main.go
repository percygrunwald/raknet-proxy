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
		Name:    "raknet-test-server",
		Usage:   "Test server that accepts connections from clients",
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

	listenAddr := fmt.Sprintf(":%d", flagValueListenPort)
	log.Debugf("listening on %v", listenAddr)
	listener, err := raknet.Listen(listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %v: %w", listenAddr, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("connection to listener failed: %v", err)
			continue
		}

		log.Tracef("client connected: %v", conn.RemoteAddr())

		conn.Close()
	}
}
