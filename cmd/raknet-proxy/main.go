package main

import (
	"os"

	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "raknet-proxy",
		Usage:   "RakNet proxy",
		Flags:   cliFlags,
		Action:  runApp,
		Version: "v0.1.5",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApp(cCtx *cli.Context) error {
	logLevel := getLogLevel(flagValueLogLevel)
	logFormat := getLogFormat(flagValueLogFormat)
	log.SetFormatter(logFormat.formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel.level)
	log.Info("Started app...")
	return nil
}