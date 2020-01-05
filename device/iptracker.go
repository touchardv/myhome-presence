package device

import (
	"encoding/binary"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ipTracker struct {
	sequenceNumber int
	devices        map[string]Device
	socket         *icmp.PacketConn
}

func newIPTracker() ipTracker {
	return ipTracker{
		sequenceNumber: 0,
		devices:        make(map[string]Device, 10),
	}
}

const data = "AreYouThere"

func (t *ipTracker) init(devices []Device) error {
	for _, device := range devices {
		targetAddr := &net.UDPAddr{IP: net.ParseIP(device.Address)}
		t.devices[targetAddr.String()] = device
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

func (t *ipTracker) waitForPingReplies(presence chan string) {
	go func() {
		now := time.Now()
		t.socket.SetReadDeadline(now.Add(15 * time.Second))
		incomingBytes := make([]byte, 32*1024)
		for {
			_, remoteAddr, err := t.socket.ReadFrom(incomingBytes)
			if err != nil {
				// check if it is a timeout error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return
				}
				log.Error("Failed reading from socket: ", err)
				return
			}
			if ipv4.ICMPType(incomingBytes[0]) != ipv4.ICMPTypeEchoReply {
				continue
			}
			sequenceNumber := binary.BigEndian.Uint16(incomingBytes[6:8])
			if sequenceNumber != uint16(t.sequenceNumber) {
				log.Warn("Ignore echo reply with wrong sequence: ", sequenceNumber, " expected: ", uint16(t.sequenceNumber))
				continue
			}
			if device, ok := t.devices[remoteAddr.String()]; ok {
				presence <- device.Identifier
			} else {
				log.Warn("Ignoring ping reply from: ", remoteAddr.String())
			}
		}
	}()
}

func (t *ipTracker) ping(devices []Device) error {
	t.sequenceNumber++
	request := icmp.Echo{ID: os.Getpid(), Seq: t.sequenceNumber, Data: []byte(data)}
	message := icmp.Message{Type: ipv4.ICMPTypeEcho, Body: &request}
	outgoingBytes, err := message.Marshal(nil)

	for _, device := range devices {
		log.Debug("Sending ping to: ", device.Description)
		targetIP := net.ParseIP(device.Address)
		targetAddr := &net.UDPAddr{IP: targetIP}
		_, err = t.socket.WriteTo(outgoingBytes, targetAddr)
		if err != nil {
			log.Warn("Ping failed: ", err)
			return err
		}
	}
	return nil
}

func (t *ipTracker) track(devices []Device, presence chan string, stopping chan struct{}) {
	err := t.init(devices)
	if err != nil {
		return
	}
	defer t.socket.Close()

	log.Info("Starting: ip tracker")
	for {
		t.waitForPingReplies(presence)
		t.ping(devices)

		ticker := time.NewTicker(1 * time.Minute)
		select {
		case <-stopping:
			log.Info("Stopped: ip tracker")
			return

		case <-ticker.C:
			break
		}
	}
}
