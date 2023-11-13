package main

import (
	"os"

	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
	_cli "github.com/urfave/cli/v2"

	"github.com/percygrunwald/raknet-proxy/lib/cli"
	"github.com/percygrunwald/raknet-proxy/lib/proxy"
)

func main() {
	app := &_cli.App{
		Name:    "raknet-proxy",
		Usage:   "RakNet proxy",
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

	proxy := &proxy.Proxy{
		ServerHostname: flagValueServerHostname,
		ServerPort:     flagValueServerPort,
		ListenPort:     flagValueListenPort,
		ProxyHostname:  flagValueProxyHostname,
	}

	return proxy.Run()
}
