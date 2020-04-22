package device

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ipTracker struct {
	sequenceNumber int
	devices        map[string]config.Device
	doneReceiving  chan bool
	socket         *icmp.PacketConn
}

func newIPTracker() Tracker {
	return &ipTracker{
		sequenceNumber: 0,
		doneReceiving:  make(chan bool),
		devices:        make(map[string]config.Device, 10),
	}
}

const pingPacketCount = 20
const pingPacketDelay = 200 * time.Millisecond
const data = "AreYouThere"

func (t *ipTracker) init(devices []config.Device) error {
	for _, device := range devices {
		if len(device.IPAddress) == 0 {
			continue
		}
		targetAddr := &net.UDPAddr{IP: net.ParseIP(device.IPAddress)}
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

func (t *ipTracker) receivePingReplies(duration time.Duration, presence chan string) {
	go func() {
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
		t.doneReceiving <- true
		log.Debug("Done receiving ping packets")
	}()
}

func (t *ipTracker) sendPingRequests() {
	t.sequenceNumber++
	request := icmp.Echo{ID: os.Getpid(), Seq: t.sequenceNumber, Data: []byte(data)}
	message := icmp.Message{Type: ipv4.ICMPTypeEcho, Body: &request}
	outgoingBytes, err := message.Marshal(nil)

	for i := 1; i <= pingPacketCount; i++ {
		log.Debug("Sending ping packet: ", i, "/", pingPacketCount)
		for _, device := range t.devices {
			targetIP := net.ParseIP(device.IPAddress)
			targetAddr := &net.UDPAddr{IP: targetIP}
			_, err = t.socket.WriteTo(outgoingBytes, targetAddr)
			if err != nil {
				msg := err.Error()
				if !strings.Contains(msg, "sendto: host is down") && !strings.Contains(msg, "no route to host") && !strings.Contains(msg, "sendto: network is unreachable") {
					log.Warn("Ping failed: ", err)
				}
			}
		}
		time.Sleep(pingPacketDelay)
	}
	log.Debug("Done sending ping packets")
}

func (t *ipTracker) Track(devices []config.Device, presence chan string, stopping chan struct{}) {
	log.Info("Starting: ip tracker")
	ticker := time.NewTicker(1 * time.Minute)
	for {
		err := t.init(devices)
		if err != nil {
			log.Error("Init failed: ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		t.receivePingReplies(15*time.Second, presence)
		t.sendPingRequests()
		<-t.doneReceiving
		t.socket.Close()

		select {
		case <-stopping:
			ticker.Stop()
			log.Info("Stopped: ip tracker")
			return

		case <-ticker.C:
			break
		}
	}
}
