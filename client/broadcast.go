package client

import (
    "encoding/binary"
    "fmt"
    "net"
)

// Calculate the interface broadcast address from the interface unicast address
func broadcastIP(addr *net.IPNet) (net.IP, error) {
    ip4 := addr.IP.To4()
    bits, size := addr.Mask.Size()

    if size != 32 && ip4 != nil {
        return nil, fmt.Errorf("Invalid IPv4 address: %v", addr)
    }

    hostBits := uint(size - bits)
    net := binary.BigEndian.Uint32([]byte(ip4))
    bcast := net | ((1 << hostBits) - 1)

    binary.BigEndian.PutUint32([]byte(ip4), bcast)

    return ip4, nil
}

// Return first broadcast address for interface
func lookupInterfaceBroadcast(name string) (net.IP, error) {
    if iface, err := net.InterfaceByName(name); err != nil {
        return nil, err
    } else if iface.Flags & net.FlagUp == 0 {
        return nil, fmt.Errorf("Interface is down: %v", name)
    } else if iface.Flags & net.FlagBroadcast == 0 {
        return nil, fmt.Errorf("Interface is not broadcast: %v", name)
    } else {
        if addrs, err := iface.Addrs(); err != nil {
            return nil, err
        } else {
            for _, ifaceAddr := range addrs {
                switch addr := ifaceAddr.(type) {
                case *net.IPNet:
                    return broadcastIP(addr)
                }
            }
        }

        return nil, fmt.Errorf("No broadcast address for interface: %v", name)
    }
}
