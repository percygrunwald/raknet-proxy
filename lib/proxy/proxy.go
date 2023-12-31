package proxy

import (
	"encoding/hex"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	ListenPort       int
	listenAddr       *net.UDPAddr
	ProxyHostname    string
	ServerHostname   string
	ServerPort       int
	serverAddr       *net.UDPAddr
	clientListenConn *net.UDPConn
}

type UDPPayload []byte

const (
	MaxUDPSize int = 65535
)

var proxyConns = make(map[int]*proxyConnection)

func (p *Proxy) Run() error {
	serverAddrString := fmt.Sprintf("%s:%d", p.ServerHostname, p.ServerPort)
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrString)
	if err != nil {
		return fmt.Errorf("unable to resolve server %v: %w", serverAddrString, err)
	}

	listenAddrString := fmt.Sprintf(":%d", p.ListenPort)
	listenAddr, err := net.ResolveUDPAddr("udp", listenAddrString)
	if err != nil {
		return fmt.Errorf("unable to resolve listen address %v: %w", listenAddrString, err)
	}

	proxyAddrString := fmt.Sprintf("%s:%d", p.ProxyHostname, p.ListenPort)
	proxyAddr, err := net.ResolveUDPAddr("udp", proxyAddrString)

	clientListenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("unable to start client listener: %w", err)
	}
	defer clientListenConn.Close()

	log.Infof("Listening on %v, proxying to %v", listenAddr, serverAddr)
	p.listenAddr = listenAddr
	p.serverAddr = serverAddr
	p.clientListenConn = clientListenConn

	b := make([]byte, MaxUDPSize)
	for {
		n, clientAddr, err := p.clientListenConn.ReadFromUDP(b)
		if err != nil {
			log.Debugf("error reading from UDP: %v", err)
			continue
		}
		payload := b[0:n]
		log.Tracef(`read %v->%v: (%d)"%s"`, clientAddr, serverAddr, n, hex.EncodeToString(payload))

		// Check if existing conn exists for client
		pConn, ok := proxyConns[clientAddr.Port]
		if !ok {
			log.Debugf("no proxy connection found for %v, starting...", clientAddr)

			// Must be `=` and not `:=` to ensure that `pConn` is not reinitialized
			pConn, err = newProxyConnection(p.clientListenConn, clientAddr, p.serverAddr, proxyAddr)
			if err != nil {
				return fmt.Errorf(`unable to start new proxy connection for %v: %w`, clientAddr, err)
			}

			proxyConns[clientAddr.Port] = pConn
		}
		log.Tracef(`writing payload from client %v to chan <- "%s"`, clientAddr, hex.EncodeToString(payload))
		pConn.payloadsFromClientChan <- payload
	}
}
