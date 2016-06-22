package discovery

import (
    "fmt"
    "net"
    "strings"
    "testing"
)

func TestBroadcastIP(t *testing.T) {
    var testBroadcastIP = []struct{
        cidr    string
        bcast   string
        err     string
    }{
        {
            cidr:  "192.168.1.10/24",
            bcast: "192.168.1.255",
        },
        {
            cidr:  "fe80::5544:33ff:fe22:1100/64",
        },
    }

    for _, test := range testBroadcastIP {
        _, ipnet, err := net.ParseCIDR(test.cidr)
        if err != nil {
            panic(err)
        }

        bcast, err := broadcastIP(ipnet)

        if err != nil {
            if test.err == "" {
                t.Errorf("broadcastIP %v: error: %v", ipnet, err)
            } else if !strings.HasPrefix(err.Error(), test.err) {
                t.Errorf("broadcastIP %v: error mismatch: %v", ipnet, err)
            }
        } else if test.bcast == "" && bcast == nil {
            t.Logf("broadcastIP %v: ok: nil", ipnet)
        } else if bcast.String() == test.bcast {
            t.Logf("broadcastIP %v: ok: %v", ipnet, bcast)
        } else {
            t.Errorf("broadcastIP %v: mismatch: %v", ipnet, bcast)
        }
    }
}

// XXX: this test assume the system will have some interfaces configured with broadcast addresses...
// XXX: would probably be better to test using fake Interface structs...
/* 
    if addrs, err := iface.Addrs(); err != nil || len(addrs) == 0 {
        panic(err)
    } else {
        t.Logf("Interface %v: %#v", iface.Name, iface)

        for _, addr := range addrs {
            t.Logf("\t%#v", addr)
        }
    } 
*/
func TestGetInterfaceBroadcast(t *testing.T) {
    var testInterfaceBroadcast = []struct{
        iface   *net.Interface
        addrs   []net.Addr
        bcast   string
        err     string
    } {
        {
            iface: &net.Interface{
                Index: 1,
                MTU:   1500,
                Name: "test-down",
                HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
                Flags: (net.FlagMulticast | net.FlagBroadcast),
            },
            err: "Interface is down: test-down",
        },
        {
            iface: &net.Interface{
                Index: 1,
                MTU:   1500,
                Name: "test-empty",
                HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
                Flags: (net.FlagMulticast | net.FlagBroadcast | net.FlagUp),  // 0x13
            },
            addrs: []net.Addr{

            },
            err: "No broadcast address for interface: test-empty",
        },
        {
            iface: &net.Interface{
                Index: 1,
                MTU:   1500,
                Name: "test-basic",
                HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
                Flags: (net.FlagMulticast | net.FlagBroadcast | net.FlagUp),  // 0x13
            },
            addrs: []net.Addr{
                // 192.168.1.10/24
                &net.IPNet{IP:net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xc0, 0xa8, 0x1, 0xa}, Mask:net.IPMask{0xff, 0xff, 0xff, 0x0}},
                // fe80::5544:33ff:fe22:1100/64
                &net.IPNet{IP:net.IP{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x55, 0x44, 0x33, 0xff, 0xfe, 0x22, 0x11, 0x00}, Mask:net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
            },
            bcast: "192.168.1.255",
        },
        {
            iface: &net.Interface{
                Index: 1,
                MTU:   1500,
                Name: "test-v6only",
                HardwareAddr: net.HardwareAddr{0x5, 0x44, 0x33, 0x22, 0x11, 0x00},
                Flags: (net.FlagMulticast | net.FlagBroadcast | net.FlagUp),  // 0x13
            },
            addrs: []net.Addr{
                // fe80::5544:33ff:fe22:1100/64
                &net.IPNet{IP:net.IP{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x55, 0x44, 0x33, 0xff, 0xfe, 0x22, 0x11, 0x00}, Mask:net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
            },
            err: "No broadcast address for interface: test-v6only",
        },
    }

    for _, test := range testInterfaceBroadcast {
        if bcast, err := getInterfaceBroadcast(test.iface, test.addrs); err != nil {
            if test.err == "" {
                t.Errorf("ERR getInterfaceBroadcast %v: error: %v", test.iface.Name, err)
            } else if !strings.HasPrefix(err.Error(), test.err) {
                t.Errorf("ERR getInterfaceBroadcast %v: error: %v", test.iface.Name, err)
            } else {
                t.Logf("OK  getInterfaceBroadcast %v: error: %v", test.iface.Name, err)
            }
        } else if bcast == nil && test.bcast == "" {
            t.Logf("OK  getInterfaceBroadcast %v: %v", test.iface.Name, bcast)
        } else if fmt.Sprintf("%v", bcast) != test.bcast {
            t.Errorf("ERR getInterfaceBroadcast %v: %v", test.iface.Name, bcast)
        } else {
            t.Logf("OK  getInterfaceBroadcast %v: %v", test.iface.Name, bcast)
        }
    }
}
