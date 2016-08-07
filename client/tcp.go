package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func (options Options) TCPClient() (*TCPClient, error) {
	if options.Address == "" {
		return nil, fmt.Errorf("No Address given")
	}

	var tcpClient = TCPClient{
		options: options,
	}

	hostAddr := net.JoinHostPort(options.Address, options.TCPPort)

	if tcpAddr, err := net.ResolveTCPAddr("tcp", hostAddr); err != nil {
		return nil, fmt.Errorf("net.ResolveTCPAddr %v: %v", hostAddr, err)
	} else {
		tcpClient.tcpAddr = tcpAddr
	}

	if tcpConn, err := net.DialTCP("tcp", nil, tcpClient.tcpAddr); err != nil {
		return nil, fmt.Errorf("net.DialTCP %v: %v", tcpClient.tcpAddr, err)
	} else {
		tcpClient.tcpConn = tcpConn
	}

	tcpClient.tcpReader = bufio.NewReader(tcpClient.tcpConn)
	tcpClient.tcpWriter = bufio.NewWriter(tcpClient.tcpConn)

	go tcpClient.run()

	return &tcpClient, nil
}

type TCPClient struct {
	options Options

	// telnet API
	tcpAddr		*net.TCPAddr
	tcpConn		*net.TCPConn
	tcpReader	*bufio.Reader
	tcpWriter	*bufio.Writer
}

func (client *TCPClient) String() string {
	return client.options.Address
}

func (client *TCPClient) send(parts []string) error {
	line := strings.Join(parts, " ") + "\n"

	if client.options.Debug {
		log.Printf("client:TCPClient %v: send: %v", client, line)
	}

	if _, err := client.tcpWriter.WriteString(line); err != nil {
		return err
	} else if err := client.tcpWriter.Flush(); err != nil {
		return err
	} else {
		return nil
	}
}

func (client *TCPClient) recv() ([]string, error) {
	if line, err := client.tcpReader.ReadString('\n'); err != nil {
		return nil, err
	} else {
		return strings.Split(strings.TrimSpace(line), " "), nil
	}
}

func (client *TCPClient) run() {
	for {
		if msg, err := client.recv(); err != nil {
			log.Printf("client:TCPClient %v: recv: %v", client, err)
			return
		} else {
			log.Printf("client:TCPClient %v: recv: %#v", client, msg)
		}
	}
}

func (client *TCPClient) command(command string, params... interface{}) error {
	var parts = []string{command}

	for _, param := range params {
		parts = append(parts, fmt.Sprintf("%v", param))
	}

	return client.send(parts)
}

func (client *TCPClient) AutoTrans(frames int) error {
	return client.command("ATRN", frames)
}

func (client *TCPClient) Cut() error {
	return client.command("ATRN")
}

func (client *TCPClient) PresetSave(preset Preset) error {
	return client.command("PRESET", "-s", preset.Sno)
}

func (client *TCPClient) PresetRecall(preset Preset) error {
	return client.command("PRESET", "-r", preset.Sno)
}

func (client *TCPClient) PresetAutoTrans(preset Preset) error {
	return client.command("PRESET", "-a", preset.Sno)
}

