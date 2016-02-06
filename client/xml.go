package client

import (
    "fmt"
    "log"
    "net"
    "time"
    "encoding/xml"
)

type xmlPacket struct {
    ID          int         `xml:"id,attr"`
    GUID        string      `xml:"GuiId"`

    // SrcMgr
    DestMgr     *DestMgr    `xml:"DestMgr"`
    PresetMgr   *PresetMgr  `xml:"PresetMgr"`
}

type xmlClient struct {
    addr            *net.TCPAddr
    conn            *net.TCPConn
    timeout         time.Duration

    readChan        chan xmlPacket
    listenChan      chan xmlPacket
}

func (client *Client) xmlClient() (*xmlClient, error) {
    xmlClient := xmlClient{
        timeout:        client.options.Timeout,

        readChan:       make(chan xmlPacket),
        listenChan:     make(chan xmlPacket),
    }

    if tcpAddr, err := net.ResolveTCPAddr("tcp4", net.JoinHostPort(client.options.Address, client.options.XMLPort)); err != nil {
        return nil, fmt.Errorf("Client.initXML: ResolveTCPAddr: %v", err)
    } else {
        xmlClient.addr = tcpAddr
    }

    if tcpConn, err := net.DialTCP("tcp4", nil, xmlClient.addr); err != nil {
        return nil, fmt.Errorf("Client.initXML: DialTCP %v: %v", xmlClient.addr,err)
    } else {
        xmlClient.conn = tcpConn
    }

    go xmlClient.reader()
    go xmlClient.run()

    return &xmlClient, nil
}

func (xmlClient *xmlClient) read() (*xmlPacket, error) {
    var packet xmlPacket

    if err := xml.NewDecoder(xmlClient.conn).Decode(&packet); err != nil {
        return nil, err
    } else {
        return &packet, nil
    }
}

func (xmlClient *xmlClient) write(packet *xmlPacket) error {
    return xml.NewEncoder(xmlClient.conn).Encode(packet)
}

func (xmlClient *xmlClient) reader() {
    defer close(xmlClient.readChan)

    for {
        if packet, err := xmlClient.read(); err != nil {
            log.Printf("xmlClient.read: %v\n", err)
            return
        } else {
            xmlClient.readChan <- *packet
        }
    }
}

func (xmlClient *xmlClient) run() {
    timer := time.NewTimer(xmlClient.timeout / 2)

    for {
        select {
        case <-timer.C:
            // TODO: ping

        case xmlPacket := <-xmlClient.readChan:
            timer.Reset(xmlClient.timeout / 2)

            xmlClient.listenChan <- xmlPacket
        }
    }
}
