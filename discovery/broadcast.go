package discovery

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Calculate the IPv4 broadcast address from the interface unicast address
func broadcastIP(addr *net.IPNet) (net.IP, error) {
	if ip4 := addr.IP.To4(); ip4 != nil {
		bits, size := addr.Mask.Size()

		hostBits := uint(size - bits)
		net := binary.BigEndian.Uint32([]byte(ip4))
		bcast := net | ((1 << hostBits) - 1)

		binary.BigEndian.PutUint32([]byte(ip4), bcast)

		return ip4, nil
	} else {
		// skip
		return nil, nil
	}
}

func getInterfaceBroadcast(iface *net.Interface, addrs []net.Addr) (net.IP, error) {
	if iface.Flags&net.FlagUp == 0 {
		return nil, fmt.Errorf("Interface is down: %v", iface.Name)
	} else if iface.Flags&net.FlagBroadcast == 0 {
		return nil, fmt.Errorf("Interface is not broadcast: %v", iface.Name)
	}

	for _, ifaceAddr := range addrs {
		switch addr := ifaceAddr.(type) {
		case *net.IPNet:
			if ip, err := broadcastIP(addr); err != nil {
				return nil, err
			} else if ip != nil {
				return ip, nil
			} else {
				// skip, probably IPv6
			}
		}
	}

	return nil, fmt.Errorf("No broadcast address for interface: %v", iface.Name)
}

// Return first IPv4 broadcast address for named interface
func lookupInterfaceBroadcast(name string) (net.IP, error) {
	if iface, err := net.InterfaceByName(name); err != nil {
		return nil, err
	} else if addrs, err := iface.Addrs(); err != nil {
		return nil, err
	} else {
		return getInterfaceBroadcast(iface, addrs)
	}
}
