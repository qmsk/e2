package discovery

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

type Packet struct {
	IP net.IP

	Hostname   string
	XMLPort    int
	Name       string
	UnitID     int
	VPCount    int
	MasterMac  string
	Version    string
	IPAddress  string
	MacAddress string
	Type       string
}

func (packet *Packet) unpackHostname(hostname string) error {
	parts := strings.Split(hostname, ":")

	if len(parts) > 0 {
		packet.Hostname = parts[0]
	}
	if len(parts) > 1 {
		if _, err := fmt.Sscanf(parts[1], "%d", &packet.XMLPort); err != nil {
			return fmt.Errorf("Invalid XMLPort=%#v: %v", parts[1], err)
		}
	}
	if len(parts) > 2 {
		packet.Name = parts[2]
	}
	if len(parts) > 3 {
		if _, err := fmt.Sscanf(parts[3], "%d", &packet.UnitID); err != nil {
			return fmt.Errorf("Invalid UnitID=%#v: %v", parts[3], err)
		}
	}
	if len(parts) > 4 {
		if _, err := fmt.Sscanf(parts[4], "%d", &packet.VPCount); err != nil {
			return fmt.Errorf("Invalid VPCount=%#v: %v", parts[4], err)
		}
	}
	if len(parts) > 5 {
		packet.MasterMac = strings.Replace(parts[5], "$", ":", -1)
	}
	if len(parts) > 6 {
		packet.Version = parts[6]
	}

	return nil
}

func (packet *Packet) unpack(addr *net.UDPAddr, data []byte) error {
	packet.IP = addr.IP

	for _, field := range bytes.Split(data, []byte{0}) {
		if len(field) == 0 {
			continue
		}

		parts := strings.SplitN(string(field), "=", 2)

		if len(parts) != 2 {
			return fmt.Errorf("Invalid field: %#v", field)
		}

		value := parts[1]

		switch parts[0] {
		case "hostname":
			if err := packet.unpackHostname(value); err != nil {
				return fmt.Errorf("Invalid hostname=%#v: %v", value, err)
			}

		case "ip-address":
			packet.IPAddress = value

		case "mac-address":
			packet.MacAddress = value

		case "type":
			packet.Type = value

		}
	}

	return nil
}
