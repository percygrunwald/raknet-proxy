package main

import (
	"fmt"
	"net"
	"os"

	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
	_cli "github.com/urfave/cli/v2"

	"github.com/percygrunwald/raknet-proxy/lib/cli"
	"github.com/percygrunwald/raknet-proxy/lib/proxy"
)

func main() {
	app := &_cli.App{
		Name:    "mirror",
		Usage:   "Mirrors UDP packets back to a client",
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

	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", flagValueListenPort))
	if err != nil {
		return fmt.Errorf("unable to resolve listen address: %w", err)
	}

	listenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("unable to start client listener: %w", err)
	}
	defer listenConn.Close()
	log.Debugf("listening for UDP packets on %v", listenAddr)

	b := make([]byte, proxy.MaxUDPSize)
	for {
		n, clientAddr, err := listenConn.ReadFromUDP(b)
		if err != nil {
			log.Debugf("error reading %v->%v: %v", clientAddr, listenAddr, err)
			continue
		}
		payload := b[0:n]
		log.Tracef(`read %v->%v: (%d)"%s"`, clientAddr, listenAddr, n, payload)

		n, _, err = listenConn.WriteMsgUDP(payload, []byte{}, clientAddr)
		if err != nil {
			log.Debugf("error writing %v->%v: %v", listenAddr, clientAddr, err)
			continue
		}
		log.Tracef(`wrote %v->%v: (%d)"%s"`, listenAddr, clientAddr, n, payload)
	}
}
