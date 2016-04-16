package client

import (
    "runtime/debug"
    "fmt"
    "io"
    "log"
    "net"
    "time"
    "encoding/xml"
)

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

    xmlResponse

    // decoding state
    reset       bool
}

func (r *xmlPacket) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    reset := xmlAttr(e, "reset")

    if reset == "yes" {
        r.reset = true

        // reset
        r.System.Reset()
    }

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
    conn            net.Conn
    timeout         time.Duration

    readChan        chan xmlPacket
    readError       error           // set once readChan is closed, nil on clean EOF
    listenChan      chan System
}

func (options Options) XMLClient() (*XMLClient, error) {
    xmlClient := XMLClient{
        timeout:        options.Timeout,
    }

    if tcpAddr, err := net.ResolveTCPAddr("tcp4", net.JoinHostPort(options.Address, options.XMLPort)); err != nil {
        return nil, fmt.Errorf("Client.initXML: ResolveTCPAddr: %v", err)
    } else if tcpConn, err := net.DialTCP("tcp4", nil, tcpAddr); err != nil {
        return nil, fmt.Errorf("Client.initXML: DialTCP %v: %v", tcpAddr, err)
    } else {
        xmlClient.conn = tcpConn
    }

    if err := xmlClient.start(); err != nil {
        return nil, err
    }

    return &xmlClient, nil
}

// Blocking read and decode of a complete <System> state into the given xmlPacket.
// Any existing xmlPacket.System state is updated, or reset if the server returns a <System reset="yes">
//
// The read is performed using a timeout deadline for the entire <System> state
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

// Blocking encode and write of an arbitrary XML packet to the server.
func (xmlClient *XMLClient) write(packet interface{}) error {
    if err := xmlClient.conn.SetWriteDeadline(time.Now().Add(xmlClient.timeout)); err != nil {
        return err
    }

    return xml.NewEncoder(xmlClient.conn).Encode(packet)
}

// Send a <System> reset query
func (xmlClient *XMLClient) writeReset() error {
    return xmlClient.write(xmlQuery{Reset: "yes",
        Type:       3,
        Query:      3,
        Recursive:  1,
    })
}

// Send a <System> ping
func (xmlClient *XMLClient) writePing() error {
    return xmlClient.write(xmlPing{})
}

// Launch background goroutines
func (xmlClient *XMLClient) start() error {
    xmlClient.listenChan = make(chan System)
    xmlClient.readChan = make(chan xmlPacket)

    go xmlClient.reader()
    go xmlClient.writer()

    return nil
}

// Read and handle messages from the server, dispatching readChan for timing and listenChan for readers
//
// Guarantees completion by closing the readChan and setting readError
func (xmlClient *XMLClient) reader() {
    defer close(xmlClient.listenChan)
    defer close(xmlClient.readChan)

    // wrap to return panics via .readError
    defer func(){
        panicValue := recover()

        if panicValue == nil {
            xmlClient.readError = nil
        } else if panicError, ok := panicValue.(Error); ok {
            xmlClient.readError = panicError
        } else {
            xmlClient.readError = fmt.Errorf("%v\n%v", panicValue, string(debug.Stack()))
        }
    }()

    var system System
    var wantReset = true

    for {
        packet := xmlPacket{xmlResponse: xmlResponse{System: &system}}

        if err := xmlClient.read(&packet); err != nil {
            log.Printf("xmlClient.read: %v\n", err)

            if err == io.EOF {
                // done
                return

            } else {
                // quit with error
                panic(err)
            }

        } else if wantReset && !packet.reset {
            // skip packets before initial reset-sync
            log.Printf("xmlClient.read: skip")

        } else {
            wantReset = false

            // timeout handling
            xmlClient.readChan <- packet

            if packet.reset {
                log.Printf("xmlClient.read: reset")

                xmlClient.listenChan <- system

            } else if packet.Type == 0 {
                log.Printf("xmlClient.read: update")

                xmlClient.listenChan <- system

            } else {
                // log.Printf("xmlClient.read: pong")
            }
        }
    }
}

// Write requests, including the initial reset to sync state, and periodic pings on idle timeout
func (xmlClient *XMLClient) writer() {
    defer xmlClient.conn.Close()

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

            // fast-retry ping on timeout
            timer.Reset(xmlClient.timeout / 8)

        case _, ok := <-xmlClient.readChan:
            if !ok {
                return
            }

            // got update, reschedule idle ping
            timer.Reset(xmlClient.timeout / 2)
        }
    }
}

// Listen for system updates.
//
// Read updated System states from the given chan, starting from the initial reset state.
//
// The received System is safe for concurrent reads, but do *not* write to it! Use the public methods if possible.
//
// XXX: only one shared chan for all callers
func (xmlClient *XMLClient) Listen() (chan System, error) {
    if xmlClient.readError != nil {
        return nil, xmlClient.readError
    } else {
        return xmlClient.listenChan, nil
    }
}

// Return any Error after Listen() returns, or nil if the connection was closed cleanly by the other end (EOF)
func (xmlClient *XMLClient) ListenError() error {
    return xmlClient.readError
}

// Read system updates.
//
// Blocking read of updated System state, or error (including io.EOF on clean shutdown).
//
// The received System is safe for concurrent reads, but do *not* write to it! Use the public methods if possible.
func (xmlClient *XMLClient) Read() (System, error) {
    if system, ok := <-xmlClient.listenChan; ok {
        return system, nil
    } else if xmlClient.readError != nil {
        // chan was closed with an error
        return system, xmlClient.readError
    } else {
        // chan was closed after EOF
        return system, io.EOF
    }
}
