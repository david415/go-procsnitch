package procsnitch

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type DummyListener struct {
	network   string
	address   string
	waitGroup *sync.WaitGroup
}

func NewDummyListener(network, address string, wg *sync.WaitGroup) *DummyListener {
	l := DummyListener{
		network:   network,
		address:   address,
		waitGroup: wg,
	}
	return &l
}

func (l *DummyListener) AcceptLoop() {
	l.waitGroup.Add(1)
	listener, err := net.Listen(l.network, l.address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	l.waitGroup.Done()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go l.SessionWorker(conn)
	}
}

func (l *DummyListener) SessionWorker(conn net.Conn) {
	for {
		time.Sleep(time.Second * 60)
	}
}

func TestLookupTCPSocketProcess(t *testing.T) {
	// listen for a connection
	var wg sync.WaitGroup
	network := "tcp"
	address := "127.0.0.1:6655"
	l := NewDummyListener(network, address, &wg)
	go l.AcceptLoop()
	wg.Wait()

	// dial a connection
	conn, err := net.Dial(network, address)
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

func TestLookupUNIXSocketProcess(t *testing.T) {
	// listen for a connection
	var wg sync.WaitGroup
	network := "unix"
	address := "testing_socket"
	l := NewDummyListener(network, address, &wg)
	go l.AcceptLoop()
	wg.Wait()

	// XXX fix me
	time.Sleep(time.Second * 1)

	// dial a connection
	conn, err := net.Dial(network, address)
	if err != nil {
		panic(err)
	}
	defer os.Remove(address)

	conn.Write([]byte("hello"))

	// prepare to do proc lookup
	fields := strings.Split(conn.RemoteAddr().String(), ":")
	fmt.Printf("remote addr %s", fields)
	fields = strings.Split(conn.LocalAddr().String(), ":")
	fmt.Printf("local addr %s", fields)
	//procInfo := LookupUNIXSocketProcess(uint16(srcP), dstIP, uint16(dstP))
	//fmt.Printf("info %s\n", procInfo)
}
