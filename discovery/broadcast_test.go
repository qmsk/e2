package discovery

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

type broadcastIPtest struct {
	cidr  string
	bcast string
	err   string
}

func testBroadcastIP(t *testing.T, test broadcastIPtest) {
	_, ipnet, err := net.ParseCIDR(test.cidr)
	if err != nil {
		panic(err)
	}

	bcast, err := broadcastIP(ipnet)

	if test.err != "" {
		assert.EqualError(t, err, test.err)
	} else if err != nil {
		assert.Errorf(t, err, "broadcastIP(%#v)", ipnet)
	} else if test.bcast == "" {
		assert.Nilf(t, bcast, "broadcastIP(%#v)", ipnet)
	} else {
		assert.Equalf(t, test.bcast, bcast.String(), "broadcastIP(%#v)", ipnet)
	}
}

func TestBroadcastIPv4(t *testing.T) {
	testBroadcastIP(t, broadcastIPtest{
		cidr:  "192.168.1.10/24",
		bcast: "192.168.1.255",
	})
}

func TestBroadcastIPv6(t *testing.T) {
	testBroadcastIP(t, broadcastIPtest{
		cidr: "fe80::5544:33ff:fe22:1100/64",
	})
}

type interfaceBroadcastTest struct {
	iface *net.Interface
	addrs []net.Addr
	bcast string
	err   string
}

func testGetInterfaceBroadcast(t *testing.T, test interfaceBroadcastTest) {
	bcast, err := getInterfaceBroadcast(test.iface, test.addrs)

	if test.err != "" {
		assert.EqualError(t, err, test.err)
	} else if err != nil {
		assert.Errorf(t, err, "getInterfaceBroadcast(%#v, %#v)", test.iface, test.addrs)
	} else if test.bcast == "" {
		assert.Nilf(t, bcast, "getInterfaceBroadcast(%#v, %#v)", test.iface, test.addrs)
	} else {
		assert.Equalf(t, test.bcast, bcast.String(), "getInterfaceBroadcast(%#v, %#v)", test.iface, test.addrs)
	}
}

func TestGetInterfaceBroadcast(t *testing.T) {
	testGetInterfaceBroadcast(t, interfaceBroadcastTest{
		iface: &net.Interface{
			Index:        1,
			MTU:          1500,
			Name:         "test-down",
			HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
			Flags:        (net.FlagMulticast | net.FlagBroadcast),
		},
		err: "Interface is down: test-down",
	})

	testGetInterfaceBroadcast(t, interfaceBroadcastTest{
		iface: &net.Interface{
			Index:        1,
			MTU:          1500,
			Name:         "test-empty",
			HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
			Flags:        (net.FlagMulticast | net.FlagBroadcast | net.FlagUp), // 0x13
		},
		addrs: []net.Addr{},
		err:   "No broadcast address for interface: test-empty",
	})

	testGetInterfaceBroadcast(t, interfaceBroadcastTest{
		iface: &net.Interface{
			Index:        1,
			MTU:          1500,
			Name:         "test-basic",
			HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
			Flags:        (net.FlagMulticast | net.FlagBroadcast | net.FlagUp), // 0x13
		},
		addrs: []net.Addr{
			// 192.168.1.10/24
			&net.IPNet{IP: net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xc0, 0xa8, 0x1, 0xa}, Mask: net.IPMask{0xff, 0xff, 0xff, 0x0}},
			// fe80::5544:33ff:fe22:1100/64
			&net.IPNet{IP: net.IP{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x55, 0x44, 0x33, 0xff, 0xfe, 0x22, 0x11, 0x00}, Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		},
		bcast: "192.168.1.255",
	})

	testGetInterfaceBroadcast(t, interfaceBroadcastTest{
		iface: &net.Interface{
			Index:        1,
			MTU:          1500,
			Name:         "test-v6only",
			HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
			Flags:        (net.FlagMulticast | net.FlagBroadcast | net.FlagUp), // 0x13
		},
		addrs: []net.Addr{
			// fe80::5544:33ff:fe22:1100/64
			&net.IPNet{IP: net.IP{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x55, 0x44, 0x33, 0xff, 0xfe, 0x22, 0x11, 0x00}, Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		},
		err: "No broadcast address for interface: test-v6only",
	})
}
