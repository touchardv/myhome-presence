package bluetooth

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	l2CAPCommandHeaderSize     = 4
	l2CAPCommandRejectResponse = 0x01
	l2CAPEchoRequest           = 0x08
	l2CAPEchoResponse          = 0x09
	l2CAPDataSize              = 20
	maxRetry                   = 5
	retryDelay                 = time.Duration(1) * time.Second
	timeout                    = 5 * 1000 // seconds
)

func ba2str(sa unix.Sockaddr) string {
	ba := sa.(*unix.SockaddrL2)
	var s strings.Builder
	for i := len(ba.Addr); i > 0; i-- {
		if i != len(ba.Addr) {
			s.WriteString(":")
		}
		s.WriteString(fmt.Sprintf("%02X", ba.Addr[i-1]))
	}
	return s.String()
}

func str2ba(addr string) unix.SockaddrL2 {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[i] = byte(u)
	}
	return unix.SockaddrL2{
		Addr: b,
		PSM:  1,
	}
}

func respondToPing(svr string) bool {
	// Create socket
	sk, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_RAW, unix.BTPROTO_L2CAP)
	if err != nil {
		log.Debug("Can't create socket: ", err)
		return false
	}
	defer unix.Close(sk)

	// Bind to local address
	bdaddrAny := unix.SockaddrL2{
		Addr: [6]uint8{0, 0, 0, 0, 0, 0}, // BDADDR_ANY
	}
	err = unix.Bind(sk, &bdaddrAny)
	if err != nil {
		log.Debug("Can't bind socket: ", err)
		return false
	}

	// Connect to the remote device
	addr := str2ba(svr)
	err = unix.Connect(sk, &addr)
	if err != nil {
		log.Debug("Can't connect: ", err)
		return false
	}

	// Get local address
	localAddr, err := unix.Getsockname(sk)
	if err != nil {
		log.Debug("Can't get local address: ", err)
		return false
	}

	str := ba2str(localAddr)
	log.Debugf("Ping: %s from %s (data size %d) ...\n", svr, str, l2CAPDataSize)

	// Initialize send buffer
	sendBuff := make([]byte, l2CAPCommandHeaderSize+l2CAPDataSize, l2CAPCommandHeaderSize+l2CAPDataSize)
	receiveBuff := make([]byte, l2CAPCommandHeaderSize+l2CAPDataSize, l2CAPCommandHeaderSize+l2CAPDataSize)
	for i := 0; i < l2CAPDataSize; i++ {
		sendBuff[l2CAPCommandHeaderSize+i] = byte(i%40 + 'A')
	}
	sendBuff[0] = l2CAPEchoRequest
	sendBuff[2] = byte(l2CAPDataSize)
	sendBuff[3] = 0

	for id := 1; id <= maxRetry; id++ {
		// Build command header
		sendBuff[1] = byte(id)

		// Send Echo Command
		_, err := unix.Write(sk, sendBuff)
		if err != nil {
			log.Debug("Write failed: ", err)
			return false
		}

		// Wait for Echo Response
		fds := make([]unix.PollFd, 1)
		fds[0].Fd = int32(sk)
		fds[0].Events = unix.POLLIN
		for {
			n, err := unix.Poll(fds, timeout)
			if err != nil {
				log.Debug("Poll failed: ", err)
				return false
			}
			if n == 0 {
				break
			}

			n, _, err = unix.Recvfrom(sk, receiveBuff, unix.MSG_WAITALL)
			if err != nil || n == 0 {
				log.Debug("Failed to receive data")
				return false
			}

			// Check for our id
			if receiveBuff[1] != byte(id) {
				continue
			}

			// Check type
			switch receiveBuff[0] {
			case l2CAPEchoResponse:
				return true
			case l2CAPCommandRejectResponse:
				log.Debug("Peer doesn't support Echo packets")
				return false
			}
		}
		time.Sleep(retryDelay)
	}
	return false
}
