package universe

import (
	"fmt"
	"log"
	"net"
	"time"
)

func udpSender(options TallyOptions, url TallyURL) (*UDPSender, error) {
	var addr = url.Addr()
	var udpSender = UDPSender{
		options: options,
	}

	log.Printf("universe:UDPSender: %v", addr)

	if udpAddr, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return nil, fmt.Errorf("ResolveUDPAddr %v: %v", addr, err)
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		return nil, fmt.Errorf("DialUDP %v: %v", udpAddr, err)
	} else {
		udpSender.udpConn = udpConn
	}

	return &udpSender, nil
}

type UDPSender struct {
	options TallyOptions

	udpConn *net.UDPConn
}

func (udpSender *UDPSender) String() string {
	return "udp://" + udpSender.udpConn.RemoteAddr().String()

}
func (udpSender *UDPSender) send(msg string) error {
	var buf = []byte(msg + string(udpSender.options.LineFormat))

	log.Printf("universe:UDPSender %v: send %v", udpSender, buf)

	if err := udpSender.udpConn.SetWriteDeadline(time.Now().Add(udpSender.options.Timeout)); err != nil {
		return err
	}

	if _, err := udpSender.udpConn.Write(buf); err != nil {
		return err
	}

	return nil
}

func (udpSender *UDPSender) Send(msg string) error {
	// swallow errors
	if err := udpSender.send(msg); err != nil {
		log.Printf("universe:UDPSender %v: send error: %v", udpSender, err)
	}

	return nil
}

func (udpSender *UDPSender) Close() error {
	return udpSender.udpConn.Close()
}
