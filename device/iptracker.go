package device

import (
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type notifier interface {
	notify(device Device, present bool)
}

type ipTracker struct {
	sequenceNumber int
	done           chan bool
	stopped        chan bool
}

func newIPTracker() ipTracker {
	return ipTracker{
		sequenceNumber: 0,
		done:           make(chan bool, 1),
		stopped:        make(chan bool, 1),
	}
}

const data = "AreYouThere"

func (t *ipTracker) ping(device Device) (bool, error) {
	sourceIP := net.ParseIP("0.0.0.0")
	socket, err := icmp.ListenPacket("udp4", sourceIP.String())
	if err != nil {
		log.Warn("Ping failed: ", err)
		return false, err
	}
	defer socket.Close()

	request := icmp.Echo{ID: os.Getpid(), Seq: t.sequenceNumber, Data: []byte(data)}
	message := icmp.Message{Type: ipv4.ICMPTypeEcho, Body: &request}
	outgoingBytes, err := message.Marshal(nil)

	targetIP := net.ParseIP(device.Address)
	targetAddr := &net.UDPAddr{IP: targetIP}
	_, err = socket.WriteTo(outgoingBytes, targetAddr)
	if err != nil {
		log.Warn("Ping failed: ", err)
		return false, err
	}
	now := time.Now()
	socket.SetReadDeadline(now.Add(5 * time.Second))
	incomingBytes := make([]byte, 32*1024)
	for {
		_, remoteAddr, err := socket.ReadFrom(incomingBytes)
		if err != nil {
			// check if it is a timeout error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return false, nil
			}
			return false, err
		}
		if remoteAddr.String() != targetAddr.String() {
			continue
		}
		if ipv4.ICMPType(incomingBytes[0]) != ipv4.ICMPTypeEchoReply {
			continue
		}
		// TODO inspect further the incoming ICMP echo reply
		return true, nil
	}
}

func (t *ipTracker) stop() {
	t.done <- true
	<-t.stopped
}

func (t *ipTracker) track(devices []Device, n notifier) {
	go func() {
		for {
			ticker := time.NewTicker(1 * time.Minute)
			select {
			case <-t.done:
				t.stopped <- true
				return

			case <-ticker.C:
				for _, device := range devices {
					present, err := t.ping(device)
					if err == nil {
						n.notify(device, present)
					}
				}
			}
		}
	}()
}
