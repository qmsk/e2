package client

import (
    "fmt"
    "log"
    "net"
    "time"
    "encoding/xml"
)

type xmlPacket struct {
    XMLName     xml.Name    `xml:"System"`

    ID          int         `xml:"id,attr"`
    GUID        *string     `xml:"GuiId"`
    Type        *int        `xml:"XMLType"`
    Resp        *int        `xml:"Resp"`

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

func (xmlClient *xmlClient) read(packet *xmlPacket) error {
    // applies to the complete XML packet read by the decoder..?
    if err := xmlClient.conn.SetReadDeadline(time.Now().Add(xmlClient.timeout)); err != nil {
        return err
    }

    if err := xml.NewDecoder(xmlClient.conn).Decode(packet); err != nil {
        return err
    } else {
        return nil
    }
}

func (xmlClient *xmlClient) write(packet xmlPacket) error {
    if err := xmlClient.conn.SetWriteDeadline(time.Now().Add(xmlClient.timeout)); err != nil {
        return err
    }

    return xml.NewEncoder(xmlClient.conn).Encode(packet)
}

func (xmlClient *xmlClient) writePing() error {
    return xmlClient.write(xmlPacket{})
}

func (xmlClient *xmlClient) reader() {
    defer close(xmlClient.readChan)

    for {
        var packet xmlPacket

        if err := xmlClient.read(&packet); err != nil {
            log.Printf("xmlClient.read: %v\n", err)
            return
        } else {
            xmlClient.readChan <- packet
        }
    }
}

func (xmlClient *xmlClient) run() {
    if xmlClient.listenChan != nil {
        defer close(xmlClient.listenChan)
    }
    timer := time.NewTimer(xmlClient.timeout / 2)

    for {
        select {
        case <-timer.C:
            if err := xmlClient.writePing(); err != nil {
                log.Printf("xmlClient.write: %v\n", err)
                return
            }

            // fast-ping
            timer.Reset(xmlClient.timeout / 8)

        case xmlPacket, ok := <-xmlClient.readChan:
            if !ok {
                return
            }

            timer.Reset(xmlClient.timeout / 2)

            if xmlClient.listenChan != nil {
                xmlClient.listenChan <- xmlPacket
            }
        }
    }
}
