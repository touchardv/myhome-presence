package ipv4

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/pkg/model"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const pingPacketCount = 5
const pingPacketDelay = 100 * time.Millisecond
const data = "AreYouThere"

func (t *ipTracker) receiveLoop(deviceReport device.ReportPresenceFunc) {
	incomingBytes := make([]byte, 32*1024)
	log.Debug("Receiving ping packets")
	for {
		n, _, remoteAddr, err := t.socket.IPv4PacketConn().ReadFrom(incomingBytes)
		if err != nil {
			// check if it is a timeout error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Debug("Timeout reading from socket")
				continue
			}
			if !t.stopReceiving {
				log.Error("Failed reading packet: ", err)
			}
			break
		}
		log.Trace("Received ", n, " bytes: ", incomingBytes[:n])
		if n < 8 {
			log.Error("Failed parsing icmp message: not enough data")
			continue
		}
		var m *icmp.Message
		if m, err = icmp.ParseMessage(1, incomingBytes[:n]); err != nil {
			log.Error("Failed parsing icmp message: ", err)
		}
		if m.Type != ipv4.ICMPTypeEchoReply {
			log.Trace("Ignore icmp message of type: ", *m)
			continue
		}
		sequenceNumber := binary.BigEndian.Uint16(incomingBytes[6:8])
		if (uint16(t.sequenceNumber) - sequenceNumber) > 2 {
			log.Trace("Ignore echo reply with wrong sequence: ", sequenceNumber, " expected: ", uint16(t.sequenceNumber))
			continue
		}
		switch addr := remoteAddr.(type) {
		case *net.UDPAddr:
			log.Debug("Got reply from: ", addr.IP.String())
			itf := model.Interface{Type: model.InterfaceUnknown, IPv4Address: addr.IP.String()}
			deviceReport(itf, nil)
		}
	}
	log.Debug("Done receiving ping packets")
}
func (t *ipTracker) Ping(devices []model.Device) {
	log.Debugf("Sending ping to %d device(s)", len(devices))
	t.sequenceNumber++
	for _, d := range devices {
		t.ping(d)
	}
}

func (t *ipTracker) ping(d model.Device) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  t.sequenceNumber,
			Data: []byte(data),
		},
	}
	outgoingBytes, _ := message.Marshal(nil)

	for i := 1; i <= pingPacketCount; i++ {
		for _, itf := range d.Interfaces {
			if itf.Type == model.InterfaceEthernet ||
				itf.Type == model.InterfaceWifi {
				log.Debugf("Sending ping packet to %s (%s) %d/%d ", d.Identifier, itf.IPv4Address, i, pingPacketCount)
				targetIP := net.ParseIP(itf.IPv4Address)
				targetAddr := &net.UDPAddr{IP: targetIP}
				_, err := t.socket.WriteTo(outgoingBytes, targetAddr)
				if err != nil {
					msg := err.Error()
					if !strings.Contains(msg, "sendto: host is down") && !strings.Contains(msg, "no route to host") && !strings.Contains(msg, "sendto: network is unreachable") {
						log.Warn("Ping failed: ", err)
					}
				}
			}
		}
		time.Sleep(pingPacketDelay)
	}
	log.Debug("Done sending ping packets")
}
