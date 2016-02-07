package client

import (
    "fmt"
    "log"
    "net"
    "time"
    "encoding/xml"
)

func xmlAttr(e xml.StartElement, name string) (value string) {
    for _, attr := range e.Attr {
        if attr.Name.Local == name {
            return attr.Value
        }
    }

    return ""
}

func xmlID(e xml.StartElement) (id int, err error) {
    value := xmlAttr(e, "id")

    if _, err := fmt.Sscanf(value, "%d", &id); err != nil {
        return id, err
    } else {
        return id, nil
    }
}

type xmlResponse struct {
    ID          int         `xml:"id,attr"`
    Reset       string      `xml:"reset,attr"`

    GUID        string      `xml:"GuiId"`
    Type        int         `xml:"XMLType"`
    Resp        int         `xml:"Resp"`

    *System
}

// Decode packet from server, based on response type.
// The given *System will be reset or updated
type xmlPacket struct {
    XMLName     xml.Name    `xml:"System"`

    reset       bool

    xmlResponse

    system      *System
}

func (r *xmlPacket) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    reset := xmlAttr(e, "reset")

    if reset == "yes" {
        r.reset = true

        // reset
        r.system.Reset()
    }

    // decode into selected *System
    r.xmlResponse.System = r.system

    return d.DecodeElement(&r.xmlResponse, &e)
}

type xmlPing struct {
    XMLName     xml.Name    `xml:"System"`
    ID          int         `xml:"id,attr"`
}

type xmlQuery struct {
    XMLName     xml.Name    `xml:"System"`
    ID          int         `xml:"id,attr"`
    Reset       string      `xml:"reset,attr"`

    Type        int         `xml:"XMLType"`
    Query       int         `xml:"Query"`
    Recursive   int         `xml:"Recursive"`
}

type XMLClient struct {
    addr            *net.TCPAddr
    conn            *net.TCPConn
    timeout         time.Duration

    readChan        chan xmlPacket
    listenChan      chan System

    // state
    system          System
}

func (options Options) XMLClient() (*XMLClient, error) {
    xmlClient := XMLClient{
        timeout:        options.Timeout,
    }

    if tcpAddr, err := net.ResolveTCPAddr("tcp4", net.JoinHostPort(options.Address, options.XMLPort)); err != nil {
        return nil, fmt.Errorf("Client.initXML: ResolveTCPAddr: %v", err)
    } else {
        xmlClient.addr = tcpAddr
    }

    if tcpConn, err := net.DialTCP("tcp4", nil, xmlClient.addr); err != nil {
        return nil, fmt.Errorf("Client.initXML: DialTCP %v: %v", xmlClient.addr,err)
    } else {
        xmlClient.conn = tcpConn
    }

    return &xmlClient, nil
}

func (xmlClient *XMLClient) read(packet *xmlPacket) error {
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

func (xmlClient *XMLClient) write(packet interface{}) error {
    if err := xmlClient.conn.SetWriteDeadline(time.Now().Add(xmlClient.timeout)); err != nil {
        return err
    }

    return xml.NewEncoder(xmlClient.conn).Encode(packet)
}

func (xmlClient *XMLClient) writeReset() error {
    return xmlClient.write(xmlQuery{Reset: "yes",
        Type:       3,
        Query:      3,
        Recursive:  1,
    })
}

func (xmlClient *XMLClient) writePing() error {
    return xmlClient.write(xmlPing{})
}

func (xmlClient *XMLClient) start() {
    if xmlClient.readChan == nil {
        xmlClient.readChan = make(chan xmlPacket)

        go xmlClient.reader()
        go xmlClient.run()
    }
}

func (xmlClient *XMLClient) reader() {
    defer close(xmlClient.readChan)

    var wantReset = true

    for {
        packet := xmlPacket{system: &xmlClient.system}

        if err := xmlClient.read(&packet); err != nil {
            log.Printf("xmlClient.read: %v\n", err)
            return
        } else if wantReset && !packet.reset {
            // skip packets before initial reset-sync
            log.Printf("xmlClient.read: skip")
        } else {
            wantReset = false

            xmlClient.readChan <- packet

            if packet.reset {
                log.Printf("xmlClient.read: reset")

                if xmlClient.listenChan != nil {
                    xmlClient.listenChan <- xmlClient.system
                }

            } else if packet.Type == 0 {
                log.Printf("xmlClient.read: update")

                if xmlClient.listenChan != nil {
                    xmlClient.listenChan <- xmlClient.system
                }

            } else {
                // log.Printf("xmlClient.read: pong")
            }
        }
    }
}

func (xmlClient *XMLClient) run() {
    if xmlClient.listenChan != nil {
        defer close(xmlClient.listenChan)
    }
    timer := time.NewTimer(xmlClient.timeout / 2)

    // initialize
    if err := xmlClient.writeReset(); err != nil {
        log.Printf("xmlClient.writeReset: %v\n", err)
        return
    }

    for {
        select {
        case <-timer.C:
            if err := xmlClient.writePing(); err != nil {
                log.Printf("xmlClient.writePing: %v\n", err)
                return
            }

            // fast-ping
            timer.Reset(xmlClient.timeout / 8)

        case _, ok := <-xmlClient.readChan:
            if !ok {
                return
            }

            timer.Reset(xmlClient.timeout / 2)
        }
    }
}

// Listen for system updates.
//
// Starts the XML stream and sends each updated System state on the given chan, starting from the initial reset state
//
// XXX: the System is not safe (modified in-place...)
// XXX: only one shared chan for all callers
func (xmlClient *XMLClient) Listen() (chan System, error) {
    if xmlClient.listenChan != nil {
        return nil, fmt.Errorf("Already listening..")
    }

    xmlClient.listenChan = make(chan System)
    xmlClient.start()

    return xmlClient.listenChan, nil
}
