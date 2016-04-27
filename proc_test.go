package procsnitch

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

type DummyListener struct {
	network string
	address string
}

func NewDummyListener(network, address string) *DummyListener {
	l := DummyListener{
		network: network,
		address: address,
	}
	return &l
}

func (l *DummyListener) AcceptLoop() {
	listener, err := net.Listen(l.network, l.address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go l.SessionWorker(conn)
	}
}

func (d *DummyListener) SessionWorker(conn net.Conn) {
	for {
		time.Sleep(time.Second * 2)
	}
}

type DummyDialer struct {
	network string
	address string
}

func NewDummyDialer(network, address string) *DummyDialer {
	d := DummyDialer{
		network: network,
		address: address,
	}
	return &d
}

func (d *DummyDialer) Dial() (net.Conn, error) {
	return net.Dial(d.network, d.address)
}

func TestLookupTCPSocketProcess(t *testing.T) {
	// listen for a connection
	network := "tcp"
	address := "127.0.0.1:6655"
	l := NewDummyListener(network, address)
	go l.AcceptLoop()

	// dial a connection
	d := NewDummyDialer(network, address)
	conn, err := d.Dial()
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("hello"))

	// prepare to do proc lookup
	fields := strings.Split(conn.RemoteAddr().String(), ":")
	dstPortStr := fields[1]
	fields = strings.Split(conn.LocalAddr().String(), ":")
	dstIP := net.ParseIP(fields[0])
	if dstIP == nil {
		conn.Close()
		panic(fmt.Sprintf("impossible error: net.ParseIP fail for: %s\n", fields[1]))
	}
	srcP, _ := strconv.ParseUint(dstPortStr, 10, 16)
	dstP, _ := strconv.ParseUint(fields[1], 10, 16)

	procInfo := LookupTCPSocketProcess(uint16(srcP), dstIP, uint16(dstP))
	fmt.Printf("info %s\n", procInfo)
}
