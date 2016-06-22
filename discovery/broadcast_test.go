package discovery

import (
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
            err:   "Invalid IPv4 address: ",
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
        } else if bcast.String() != test.bcast {
            t.Errorf("broadcastIP %v: mismatch: %v", ipnet, bcast)
        }
    }
}
