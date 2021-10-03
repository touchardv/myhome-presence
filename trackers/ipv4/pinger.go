package ipv4

import (
	"encoding/binary"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const pingPacketCount = 5
const pingPacketDelay = 100 * time.Millisecond
const data = "AreYouThere"

func (t *ipTracker) init(devices map[string]model.Device) error {
	for _, device := range devices {
		for _, itf := range device.Interfaces {
			if itf.Type == model.InterfaceEthernet ||
				itf.Type == model.InterfaceWifi {
				targetAddr := &net.UDPAddr{IP: net.ParseIP(itf.IPv4Address)}
				t.devices[targetAddr.String()] = device
			}
		}
	}

	sourceIP := net.ParseIP("0.0.0.0")
	socket, err := icmp.ListenPacket("udp4", sourceIP.String())
	if err != nil {
		log.Error("Failed to create UDP/ICMP socket: ", err)
		return err
	}
	t.socket = socket
	return nil
}

func (t *ipTracker) receivePingReplies(devices map[string]model.Device, duration time.Duration, presence chan string) {
	go func() {
		offset := 0
		if runtime.GOOS == "darwin" {
			offset = 20
		}
		now := time.Now()
		t.socket.SetReadDeadline(now.Add(duration))
		incomingBytes := make([]byte, 32*1024)
		for {
			_, remoteAddr, err := t.socket.ReadFrom(incomingBytes)
			if err != nil {
				// check if it is a timeout error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					break
				}
				log.Error("Failed reading from socket: ", err)
				break
			}
			if ipv4.ICMPType(incomingBytes[offset+0]) != ipv4.ICMPTypeEchoReply {
				continue
			}
			sequenceNumber := binary.BigEndian.Uint16(incomingBytes[offset+6 : offset+8])
			if sequenceNumber != uint16(t.sequenceNumber) {
				log.Warn("Ignore echo reply with wrong sequence: ", sequenceNumber, " expected: ", uint16(t.sequenceNumber))
				continue
			}
			if device, ok := t.devices[remoteAddr.String()]; ok {
				log.Debug("Got reply from: ", device.Identifier, " on ", remoteAddr.String())
				presence <- device.Identifier
				delete(devices, device.Identifier)
			} else {
				log.Debug("Ignoring: ", remoteAddr.String())
			}
		}
		t.doneReceiving <- true
		log.Debug("Done receiving ping packets")
	}()
}

func (t *ipTracker) sendPingRequests(devices map[string]model.Device) {
	t.sequenceNumber++
	request := icmp.Echo{ID: os.Getpid(), Seq: t.sequenceNumber, Data: []byte(data)}
	message := icmp.Message{Type: ipv4.ICMPTypeEcho, Body: &request}
	outgoingBytes, err := message.Marshal(nil)

	for i := 1; i <= pingPacketCount; i++ {
		for _, device := range devices {
			for _, itf := range device.Interfaces {
				if itf.Type != model.InterfaceEthernet &&
					itf.Type != model.InterfaceWifi {
					continue
				}
				log.Debugf("Sending ping packet to %s (%s) %d/%d ", device.Identifier, itf.IPv4Address, i, pingPacketCount)
				targetIP := net.ParseIP(itf.IPv4Address)
				targetAddr := &net.UDPAddr{IP: targetIP}
				_, err = t.socket.WriteTo(outgoingBytes, targetAddr)
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

func (t *ipTracker) Ping(devices map[string]model.Device, presence chan string) {
	err := t.init(devices)
	if err != nil {
		log.Error("Init failed: ", err)
		return
	}
	defer t.socket.Close()

	t.receivePingReplies(devices, 10*time.Second, presence)
	t.sendPingRequests(devices)
	<-t.doneReceiving
}
