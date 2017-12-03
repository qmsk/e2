package discovery

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"testing"
	"time"
)

type discoveryTest struct {
	mock.Mock

	udpConn *net.UDPConn
	stopped bool

	discovery *Discovery
}

func (test *discoveryTest) mockResponse(b []byte) *mock.Call {
	return test.On("discover").Return(b)
}

func (test *discoveryTest) handle() []byte {
	args := test.MethodCalled("discover")

	if ret := args.Get(0); ret == nil {
		return nil
	} else {
		return ret.([]byte)
	}
}

func (test *discoveryTest) run(t *testing.T) {
	for {
		var buf = make([]byte, 1500)

		if recvSize, recvAddr, err := test.udpConn.ReadFromUDP(buf); err != nil {
			if test.stopped {
				return
			} else {
				t.Fatalf("udpConn.ReadFromUDP: %v", err)
			}
		} else if !bytes.Equal(buf[:recvSize], discoveryProbe) {
			t.Errorf("recv unknown discovery probe: %#v", buf[:recvSize])
		} else if sendBuf := test.handle(); sendBuf == nil {

		} else if _, err := test.udpConn.WriteToUDP(sendBuf, recvAddr); err != nil {
			t.Fatalf("udpConn.WriteToUDP: %v", err)
		}
	}
}

func (test *discoveryTest) stop() {
	test.stopped = true
	test.udpConn.Close()
}

func withDiscoveryTest(t *testing.T, f func(*discoveryTest)) {
	var test discoveryTest

	if udpAddr, err := net.ResolveUDPAddr("udp", "localhost:40961"); err != nil {
		panic(err)
	} else if udpConn, err := net.ListenUDP("udp", udpAddr); err != nil {
		panic(err)
	} else {
		test.udpConn = udpConn
	}

	var options = Options{
		Address:  "localhost",
		Interval: 10 * time.Millisecond,
	}

	if discovery, err := options.Discovery(); err != nil {
		panic(err)
	} else {
		test.discovery = discovery
	}

	t.Logf("Run on %v...", test.udpConn.LocalAddr())

	go test.run(t)
	defer test.stop()

	time.AfterFunc(50*time.Millisecond, func() {
		t.Fatalf("Timeout")
		test.discovery.Stop()
	})

	f(&test)
}

func TestDiscovery(t *testing.T) {
	withDiscoveryTest(t, func(test *discoveryTest) {
		test.mockResponse(testPacketBytes)

		var count = 0

		for packet := range test.discovery.Run() {
			testPacket.IP = net.IP{127, 0, 0, 1}

			assert.Equal(t, testPacket, packet)

			count++

			if count >= 2 {
				test.discovery.Stop()
			}
		}

		assert.NoError(t, test.discovery.Error())
		test.AssertNumberOfCalls(t, "discover", 2)
	})
}
