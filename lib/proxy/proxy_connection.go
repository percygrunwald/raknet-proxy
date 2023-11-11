package proxy

import (
	"encoding/binary"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

type proxyConnection struct {
	payloadsFromServerChan chan UDPPayload
	payloadsFromClientChan chan UDPPayload

	clientListenConn *net.UDPConn
	serverConn       *net.UDPConn

	clientAddr *net.UDPAddr
	serverAddr *net.UDPAddr

	clientAddrBytes []byte
	serverAddrBytes []byte
}

func newProxyConnection(clientListenConn *net.UDPConn, clientAddr *net.UDPAddr, serverAddr *net.UDPAddr) (*proxyConnection, error) {
	log.Debugf("starting proxy connection for client %v...", clientAddr)

	clientPortBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(clientPortBytes, uint16(clientAddr.Port))
	clientAddrBytes := append(clientAddr.IP, clientPortBytes...)

	serverPortBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(serverPortBytes, uint16(serverAddr.Port))
	serverAddrBytes := append(serverAddr.IP, serverPortBytes...)

	pConn := &proxyConnection{
		payloadsFromServerChan: make(chan UDPPayload, 1),
		payloadsFromClientChan: make(chan UDPPayload, 1),
		clientListenConn:       clientListenConn,
		clientAddr:             clientAddr,
		serverAddr:             serverAddr,
		clientAddrBytes:        clientAddrBytes,
		serverAddrBytes:        serverAddrBytes,
	}

	pConn.log(log.Debug, `connecting to server...`)
	go pConn.run()

	return pConn, nil
}

func (pConn *proxyConnection) logf(fn func(string, ...interface{}), msg string, args ...interface{}) {
	msg = fmt.Sprintf("[%v] %s", pConn.clientAddr, msg)
	fn(msg, args...)
}

func (pConn *proxyConnection) log(fn func(...interface{}), msg string) {
	msg = fmt.Sprintf("[%v] %s", pConn.clientAddr, msg)
	fn(msg)
}

func (pConn *proxyConnection) run() {
	pConn.logf(log.Tracef, "dialing %v...", pConn.serverAddr)

	serverConn, err := net.DialUDP("udp", nil, pConn.serverAddr)
	if err != nil {
		pConn.logf(log.Fatalf, "unable to dial upstream server UDP: %v", err)
	}
	defer serverConn.Close()
	pConn.logf(log.Tracef, "got connection to server %v->%v", serverConn.LocalAddr(), serverConn.RemoteAddr())
	pConn.serverConn = serverConn

	pConn.log(log.Debug, `starting client payload listener...`)
	go pConn.handlePayloadsFromClient()

	pConn.log(log.Debug, `starting server payload listener...`)
	go pConn.handlePayloadsFromServer()

	b := make([]byte, MaxUDPSize)
	for {
		n, _, err := serverConn.ReadFromUDP(b)
		if err != nil {
			pConn.logf(log.Debugf, "error reading %v->%v: %v", serverConn.RemoteAddr(), serverConn.LocalAddr(), err)
			continue
		}
		payload := b[0:n]
		pConn.logf(log.Tracef, `read %v->%v: (%d)"%s"`, serverConn.RemoteAddr(), serverConn.LocalAddr(), n, payload)
		pConn.logf(log.Tracef, `writing payload from server to chan <- "%s"`, payload)
		pConn.payloadsFromServerChan <- payload
	}
}

func (pConn *proxyConnection) handlePayloadsFromClient() {
	pConn.log(log.Debug, "listening for payloads from client...")

	for payload := range pConn.payloadsFromClientChan {
		pConn.logf(log.Tracef, `proxying payload from client: "%s"`, payload)
		pConn.proxyPayloadFromClient(payload)
	}
}

func (pConn *proxyConnection) handlePayloadsFromServer() {
	pConn.log(log.Debug, "listening for payloads from server...")

	for payload := range pConn.payloadsFromServerChan {
		pConn.logf(log.Tracef, `proxying payload from server: "%s"`, payload)
		pConn.proxyPayloadFromServer(payload)
	}
}

func (pConn *proxyConnection) proxyPayloadFromClient(payload UDPPayload) (int, error) {
	_ = pConn.updatePayloadFromClient(payload)
	pConn.logf(log.Tracef, `write %v->%v: "%s"`, pConn.clientAddr, pConn.serverAddr, payload)
	return pConn.serverConn.Write(payload)
}

func (pConn *proxyConnection) proxyPayloadFromServer(payload UDPPayload) (int, error) {
	_ = pConn.updatePayloadFromServer(payload)
	pConn.logf(log.Tracef, `write %v->%v: "%s"`, pConn.serverAddr, pConn.clientAddr, payload)
	n, _, err := pConn.clientListenConn.WriteMsgUDP(payload, []byte{}, pConn.clientAddr)
	return n, err
}

func (pConn *proxyConnection) updatePayloadFromServer(payload UDPPayload) error {
	switch payload[0] {
	case packetOpenConnectionReply2:
		return pConn.updatePacketOpenConnectionReply2(payload)
	case packetConnectionRequestAccepted:
		return pConn.updatePacketConnectionRequestAccepted(payload)
	default:
		return nil
	}
}

func (pConn *proxyConnection) updatePayloadFromClient(payload UDPPayload) error {
	switch payload[0] {
	case packetOpenConnectionRequest2:
		return pConn.updatePacketOpenConnectionRequest2(payload)
	case packetNewIncomingConnection:
		return pConn.updatePacketNewIncomingConnection(payload)
	default:
		return nil
	}
}

// https://wiki.vg/Raknet_Protocol#Packets
// Name						Size (b)	Range			Notes
// byte						1					0 to 255
// Long						8					-2^63 to 2^63-1	Signed 64-bit Integer
// Magic					16				00ffff00fefefefefdfdfdfd12345678	Always those hex bytes, corresponding to RakNet's default OFFLINE_MESSAGE_DATA_ID
// short					2					-32768 to 32767
// unsigned short	2					0 to 65535
// string					unsigned short + string	N/A	Prefixed by a short containing the length of the string in characters. It appears that only the following ASCII characters can be displayed: !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~
// boolean				1					0 to 1		True is encoded as 0x01, false as 0x00.
// address				7 or 29		N/A				1 byte for the IP version (4 or 6), followed by (for IPv4) 4 bytes for the IP and an unsigned short for the port number or (for IPv6) an unsigned short for the address family (always 0x17), an unsigned short for the port, 8 bytes for the flow info and 16 address bytes
// uint24le				3					N/A				3-byte little-endian unsigned integer

// Client to server
func (pConn *proxyConnection) updatePacketOpenConnectionRequest2(payload UDPPayload) error {
	// Magic					MAGIC		payload[1:7]
	// Server Address	address	payload[8] is ip version, payload[9:14] ip4 addr, payload[9:36] ip6 addr
	// MTU						short
	// Client GUID		Long
	if payload[8] == ipv4 {
		// Replace payload[9:14] with the server address and port
		copy(payload[9:14], pConn.serverAddrBytes)
	}
	return nil
}

// Client to server
func (pConn *proxyConnection) updatePacketNewIncomingConnection(payload UDPPayload) error {
	// Server address		address address	payload[1] is ip version, payload[2:8] is ip4 addr, payload[2:30] is ip6 addr
	// Internal address	address	(unknown what this is used for)
	if payload[1] == ipv4 {
		// Replace payload[2:8] with server ip and port
		copy(payload[2:8], pConn.serverAddrBytes)
	}
	return nil
}

// Server to client
func (pConn *proxyConnection) updatePacketOpenConnectionReply2(payload UDPPayload) error {
	// Magic								MAGIC		payload[1:8]
	// Server GUID					Long		payload[8:16]
	// Client Address				address	payload[16] is ip version, payload[17:23] ip4 addr, payload[17:55] ip6 addr
	// MTU									short
	// Encryption enabled?	boolean
	if payload[16] == ipv4 {
		// Replace payload[17:23] with client ip and port
		copy(payload[17:23], pConn.clientAddrBytes)
	}
	return nil
}

// Server to client
func (pConn *proxyConnection) updatePacketConnectionRequestAccepted(payload UDPPayload) error {
	// Client address		address	payload[1] is ip version, payload[2:8] is ip4 addr, payload[2:30] is ip6 addr
	// System index			short
	// Internal IDs			10x address (unknown what this is used for)
	// Request time			Long
	// Time							Long
	if payload[1] == ipv4 {
		// Replace payload[2:8] with client ip and port
		copy(payload[2:8], pConn.clientAddrBytes)
	}
	return nil
}
