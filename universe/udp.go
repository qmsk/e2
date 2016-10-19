package universe

import (
	"fmt"
	"log"
	"net"
)

func makeUDP(options TallyOptions, url TallyURL) (*UDPTally, error) {
	var addr = url.Addr()
	var udpTally = UDPTally{
		options: options,
	}

	log.Printf("universe:UDPTally: %v", addr)

	if udpAddr, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return nil, fmt.Errorf("ResolveUDPAddr %v: %v", addr, err)
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		return nil, fmt.Errorf("DialUDP %v: %v", udpAddr, err)
	} else {
		udpTally.udpConn = udpConn
	}

	return &udpTally, nil
}

type UDPTally struct {
	options TallyOptions

	udpConn *net.UDPConn
}

func (udpTally *UDPTally) String() string {
	return udpTally.udpConn.RemoteAddr().String()
}

func (udpTally *UDPTally) Send(msg string) error {
	var buf = []byte(msg + string(udpTally.options.LineFormat))

	log.Printf("universe:UDPTally %v: send %v", udpTally, buf)

	if _, err := udpTally.udpConn.Write(buf); err != nil {
		return err
	}

	return nil
}

func (udpTally *UDPTally) Close() error {
	return udpTally.udpConn.Close()
}
