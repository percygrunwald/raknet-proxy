package proxy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type proxyConnection struct {
	payloadsFromServerChan chan UDPPayload
	payloadsFromClientChan chan UDPPayload

	clientListenConn *net.UDPConn
	serverConn       *net.UDPConn

	clientAddr        *net.UDPAddr
	serverAddr        *net.UDPAddr
	proxyAsServerAddr *net.UDPAddr
	proxyAsClientAddr net.Addr

	clientAddrBytes        []byte
	serverAddrBytes        []byte
	proxyAsServerAddrBytes []byte
	proxyAsClientAddrBytes []byte
}

func newProxyConnection(clientListenConn *net.UDPConn, clientAddr *net.UDPAddr,
	serverAddr *net.UDPAddr, proxyAsServerAddr *net.UDPAddr) (*proxyConnection, error) {

	log.Debugf("starting proxy connection for client %v...", clientAddr)

	clientAddrBytes := getUDPAddrBytes(clientAddr)
	serverAddrBytes := getUDPAddrBytes(serverAddr)
	proxyAsServerAddrBytes := getUDPAddrBytes(proxyAsServerAddr)

	pConn := &proxyConnection{
		payloadsFromServerChan: make(chan UDPPayload, 1),
		payloadsFromClientChan: make(chan UDPPayload, 1),
		clientListenConn:       clientListenConn,
		clientAddr:             clientAddr,
		serverAddr:             serverAddr,
		proxyAsServerAddr:      proxyAsServerAddr,
		clientAddrBytes:        clientAddrBytes,
		serverAddrBytes:        serverAddrBytes,
		proxyAsServerAddrBytes: proxyAsServerAddrBytes,
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
	pConn.proxyAsClientAddr = serverConn.LocalAddr()
	pConn.proxyAsClientAddrBytes = getProxyAsClientAddrBytes(pConn.proxyAsServerAddr, pConn.proxyAsClientAddr)

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
		pConn.logf(log.Tracef, `read %v->%v: (%d)"%s"`, serverConn.RemoteAddr(), serverConn.LocalAddr(), n, hex.EncodeToString(payload))
		pConn.logf(log.Tracef, `writing payload from server to chan <- "%s"`, hex.EncodeToString(payload))
		pConn.payloadsFromServerChan <- payload
	}
}

func getUDPAddrBytes(addr *net.UDPAddr) []byte {
	return getIPPortBytes(addr.IP, addr.Port)
}

// Gets the byteslice representation of the address, with the port coming from
// the address, but the IP coming from the first argument
func getProxyAsClientAddrBytes(udpAddr *net.UDPAddr, addr net.Addr) []byte {
	split := strings.Split(addr.String(), ":")
	port, _ := strconv.Atoi(split[1])
	ip := udpAddr.IP

	return getIPPortBytes(ip, port)
}

// Get the byte sequence of the IP and port in the RakNet form, including the IP
// version byte (set to 4 by default)
func getIPPortBytes(ip net.IP, port int) []byte {
	// Initialize ipv4 byteslice
	addrBytes := []byte{ipv4}
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	for _, b := range ip.To4() {
		addrBytes = append(addrBytes, ^b)
	}
	addrBytes = append(addrBytes, portBytes...)

	return addrBytes
}

func (pConn *proxyConnection) handlePayloadsFromClient() {
	pConn.log(log.Debug, "listening for payloads from client...")

	for payload := range pConn.payloadsFromClientChan {
		pConn.logf(log.Tracef, `proxying payload from client: "%s"`, hex.EncodeToString(payload))
		pConn.proxyPayloadFromClient(payload)
	}
}

func (pConn *proxyConnection) handlePayloadsFromServer() {
	pConn.log(log.Debug, "listening for payloads from server...")

	for payload := range pConn.payloadsFromServerChan {
		pConn.logf(log.Tracef, `proxying payload from server: "%s"`, hex.EncodeToString(payload))
		pConn.proxyPayloadFromServer(payload)
	}
}

func (pConn *proxyConnection) proxyPayloadFromClient(payload UDPPayload) (int, error) {
	payload, _ = pConn.updatePayloadFromClient(payload)
	pConn.logf(log.Tracef, `write %v->%v: "%s"`, pConn.clientAddr, pConn.serverAddr, hex.EncodeToString(payload))
	return pConn.serverConn.Write(payload)
}

func (pConn *proxyConnection) proxyPayloadFromServer(payload UDPPayload) (int, error) {
	payload, _ = pConn.updatePayloadFromServer(payload)
	pConn.logf(log.Tracef, `write %v->%v: "%s"`, pConn.serverAddr, pConn.clientAddr, hex.EncodeToString(payload))
	n, _, err := pConn.clientListenConn.WriteMsgUDP(payload, []byte{}, pConn.clientAddr)
	return n, err
}

func (pConn *proxyConnection) updatePayloadFromServer(payload UDPPayload) (UDPPayload, error) {
	payload = bytes.ReplaceAll(payload, pConn.serverAddrBytes, pConn.proxyAsServerAddrBytes)
	payload = bytes.ReplaceAll(payload, pConn.proxyAsClientAddrBytes, pConn.clientAddrBytes)

	return payload, nil
}

func (pConn *proxyConnection) updatePayloadFromClient(payload UDPPayload) (UDPPayload, error) {
	payload = bytes.ReplaceAll(payload, pConn.clientAddrBytes, pConn.proxyAsClientAddrBytes)
	payload = bytes.ReplaceAll(payload, pConn.proxyAsServerAddrBytes, pConn.serverAddrBytes)

	return payload, nil
}
