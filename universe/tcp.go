package universe

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func tcpSender(options TallyOptions, url TallyURL) (*TCPSender, error) {
	var addr = url.Addr()
	var tcpSender = TCPSender{
		options:  options,
		sendChan: make(chan []byte, options.SendBuffer),
	}

	if tcpAddr, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr %v: %v", addr, err)
	} else {
		tcpSender.tcpAddr = tcpAddr
	}

	log.Printf("universe:TCPSender: %v", &tcpSender)

	tcpSender.runWG.Add(1)
	go tcpSender.run()

	return &tcpSender, nil
}

// The TCP connection is uni-directional. It is only used to send commands,
// and we do not expect to ever receive any commands.
// TODO: consider net:TCPConn.CloseRead?
type TCPSender struct {
	options TallyOptions

	tcpAddr *net.TCPAddr
	tcpConn *net.TCPConn
	err     error

	sendChan chan []byte
	runWG    sync.WaitGroup
}

func (tcpSender *TCPSender) String() string {
	return "tcp://" + tcpSender.tcpAddr.String()
}

func (tcpSender *TCPSender) connect() error {
	if tcpConn, err := net.DialTCP("tcp", nil, tcpSender.tcpAddr); err != nil {
		return fmt.Errorf("DialTCP %v: %v", tcpSender.tcpAddr, err)
	} else {
		tcpSender.tcpConn = tcpConn
	}

	return nil
}

func (tcpSender *TCPSender) send(msg []byte) error {
	log.Printf("universe:TCPSender %v: send msg=%#v", tcpSender, string(msg))

	if err := tcpSender.tcpConn.SetWriteDeadline(time.Now().Add(tcpSender.options.Timeout)); err != nil {
		return err
	}

	if _, err := tcpSender.tcpConn.Write(msg); err != nil {
		return err
	}

	return nil
}

func (tcpSender *TCPSender) close() error {
	tcpSender.err = tcpSender.tcpConn.Close()

	tcpSender.tcpConn = nil

	return tcpSender.err
}

func (tcpSender *TCPSender) run() {
	defer tcpSender.runWG.Done()
	defer tcpSender.close()

	// TODO: flush messages on reconnect...?
	for msg := range tcpSender.sendChan {
		if tcpSender.tcpConn != nil {

		} else if err := tcpSender.connect(); err != nil {
			tcpSender.err = err
			log.Printf("universe:TCPSender %v: drop connect: %v", tcpSender, err)
			continue
		}

		if err := tcpSender.send(msg); err != nil {
			log.Printf("universe:TCPSender %v: drop send: %v", tcpSender, err)

			tcpSender.close()
			tcpSender.err = err
		}
	}
}

func (tcpSender *TCPSender) Send(msg []byte) error {
	select {
	case tcpSender.sendChan <- msg:
		return nil
	default:
		log.Printf("universe:TCPSender %v: send dropped", tcpSender)
	}

	return nil
}

func (tcpSender *TCPSender) Close() error {
	close(tcpSender.sendChan)

	tcpSender.runWG.Wait()

	return tcpSender.err
}
