package proxy

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	ListenPort     int
	ServerHostname string
	ServerPort     int
	ProxyHostname  string
}

type UDPPayload []byte

const (
	maxUDPSize                      int  = 65535
	packetOpenConnectionRequest2    byte = 7
	packetOpenConnectionReply2      byte = 8
	packetConnectionRequestAccepted byte = 10
	packetNewIncomingConnection     byte = 13
	ipv4                            byte = 4
)

var proxyConns = make(map[int]*proxyConnection)

func (p *Proxy) Run() error {
	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", p.ListenPort))
	if err != nil {
		return fmt.Errorf("unable to resolve listen address: %w", err)
	}

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p.ServerHostname, p.ServerPort))
	if err != nil {
		return fmt.Errorf("unable to resolve upstream address: %w", err)
	}

	clientListenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("unable to start client listener: %w", err)
	}
	defer clientListenConn.Close()

	log.Infof("Listening on %v, proxying to %v", listenAddr, serverAddr)

	b := make([]byte, maxUDPSize)
	for {
		n, clientAddr, err := clientListenConn.ReadFromUDP(b)
		if err != nil {
			log.Debugf("error reading from UDP: %v", err)
			continue
		}
		payload := b[0:n]
		log.Tracef(`payload from client: n: %d, clientAddr: %v, payload: "%s"`, n, clientAddr, payload)

		// Check if existing conn exists for client
		pConn, ok := proxyConns[clientAddr.Port]
		if !ok {
			log.Debugf("no proxy connection found for clientAddr %v, starting...", clientAddr)
			// proxyAddrToServerString := fmt.Sprintf("%s:%d", p.ProxyHostname, clientAddr.Port)
			// proxyAddrToServer, err := net.ResolveUDPAddr("udp", proxyAddrToServerString)
			// if err != nil {
			// 	return fmt.Errorf(`unable to resolve proxy address "%s": %w`, proxyAddrToServerString, err)
			// }

			// proxyAddrToClientString := fmt.Sprintf("%s:%d", p.ProxyHostname, p.ServerPort)
			// proxyAddrToClient, err := net.ResolveUDPAddr("udp", proxyAddrToClientString)
			// if err != nil {
			// 	return fmt.Errorf(`unable to resolve proxy address "%s": %w`, proxyAddrToClientString, err)
			// }

			// Must be `=` and not `:=` to ensure that `pConn` is not reinitialized
			pConn, err = newProxyConnection(clientAddr, serverAddr)
			if err != nil {
				return fmt.Errorf(`unable to start new proxy connection: %w`, err)
			}

			go pConn.run()
			proxyConns[clientAddr.Port] = pConn
		}

		log.Tracef(`writing payload to client chan: clientAddr: %v, payload: "%s"`, pConn.clientAddr, payload)
		pConn.payloadsFromClientChan <- payload
	}
}
