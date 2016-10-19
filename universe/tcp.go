package universe

import (
	"fmt"
	"log"
	"net"
)

func tcpSender(options TallyOptions, url TallyURL) (*TCPSender, error) {
	var addr = url.Addr()
	var tcpSender = TCPSender{
		options: options,
	}

	log.Printf("universe:TCPSender: %v", addr)

	if tcpAddr, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr %v: %v", addr, err)
	} else if tcpConn, err := net.DialTCP("tcp", nil, tcpAddr); err != nil {
		return nil, fmt.Errorf("DialTCP %v: %v", tcpAddr, err)
	} else {
		tcpSender.tcpConn = tcpConn
	}

	return &tcpSender, nil
}

type TCPSender struct {
	options TallyOptions

	tcpConn *net.TCPConn
}

func (tcpSender *TCPSender) String() string {
	return tcpSender.tcpConn.RemoteAddr().String()
}

func (tcpSender *TCPSender) Send(msg string) error {
	var buf = []byte(msg + string(tcpSender.options.LineFormat))

	log.Printf("universe:TCPSender %v: send %v", tcpSender, buf)

	if _, err := tcpSender.tcpConn.Write(buf); err != nil {
		return err
	}

	return nil
}

func (tcpSender *TCPSender) Close() error {
	return tcpSender.tcpConn.Close()
}
