package dcp

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	// "time"
	"encoding/xml"
)

type xmlPacket struct {
	XMLName xml.Name `xml:"dcp-xml"`

	Error  string  `xml:"error,omitempty"`
	Device *Device `xml:"device"`
}

type Options struct {
	SerialName string `long:"serial-name" value-name:"/dev/tty*"`
	SerialBaud int    `long:"serial-baud" value-name:"BAUD" default:"57600"`
	// SerialTimeout   time.Duration   `long:"serial-timeout" value-name:"DURATION" default:"1s"`
}

func (options Options) Client() (*Client, error) {
	client := Client{
		options: options,
	}

	serialConfig := serial.Config{
		Name: options.SerialName,
		Baud: options.SerialBaud,
		// ReadTimeout:    options.SerialTimeout,
	}

	if serialPort, err := serial.OpenPort(&serialConfig); err != nil {
		return nil, fmt.Errorf("serial.OpenPort: %v", err)
	} else {
		client.serialPort = serialPort
	}

	client.serialReader = bufio.NewReader(client.serialPort)

	if err := client.start(); err != nil {
		return nil, err
	}

	return &client, nil
}

type Client struct {
	options      Options
	serialPort   *serial.Port
	serialReader *bufio.Reader

	state Device
}

func (client *Client) read(packet *xmlPacket) error {
	return xml.NewDecoder(client.serialReader).Decode(packet)
}

func (client *Client) write(packet interface{}) error {
	return xml.NewEncoder(client.serialPort).Encode(packet)
}

// Only works if the KVM is in Command mode
func (client *Client) query() error {
	// XXX: must be self-closing, which go's xml.Encoder does not know how to emit
	if _, err := client.serialPort.Write([]byte("<dcp-xml/>\r\n")); err != nil {
		return err
	}

	return nil
}

func (client *Client) start() error {
	return client.query()
}

func (client *Client) Read() (Device, error) {
	packet := xmlPacket{
		Device: &client.state,
	}

	if err := client.read(&packet); err != nil {
		return client.state, err
	}

	if packet.Error != "" {
		return client.state, fmt.Errorf("%s", packet.Error)
	}

	return client.state, nil
}
