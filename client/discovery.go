package client

import (
    "log"
    "time"
    "net"
)

const DISCOVERY_ADDR = "255.255.255.255"
const DISCOVERY_PORT = "40961"
const DISCOVERY_SEND = "\x3f\x00"

type DiscoveryOptions struct {
    Address         string              `long:"discovery-address" default:""`
    Interface       string              `long:"discovery-interface"`

    Interval        time.Duration       `long:"discovery-interval" default:"10s"`
}

func (options DiscoveryOptions) Discovery() (*Discovery, error) {
    discovery := &Discovery{
        options:        options,

        recvChan:       make(chan DiscoveryPacket),
    }

    if udpConn, err := net.ListenUDP("udp4", nil); err != nil {
        return nil, err
    } else {
        discovery.udpConn = udpConn
    }

    if options.Address != "" {

    } else if options.Interface != "" {
        if ip, err := lookupInterfaceBroadcast(options.Interface); err != nil {
            return nil, err
        } else {
            options.Address = ip.String()

            log.Printf("Discovery: using interface %v broadcast address: %v\n", options.Interface, options.Address)
        }
    } else {
        options.Address = DISCOVERY_ADDR
    }

    if udpAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(options.Address, DISCOVERY_PORT)); err != nil {
        return nil, err
    } else {
        discovery.udpAddr = udpAddr
    }

    return discovery, nil
}

type DiscoveryPacket struct {
    addr        *net.UDPAddr
    data        []byte

    IP          net.IP
}

type Discovery struct {
    options     DiscoveryOptions
    udpConn     *net.UDPConn
    udpAddr     *net.UDPAddr

    recvChan    chan DiscoveryPacket
}

func (discovery *Discovery) send() error {
    pkt := ([]byte)(DISCOVERY_SEND)

    if _, err := discovery.udpConn.WriteToUDP(pkt, discovery.udpAddr); err != nil {
        return err
    }

    return nil
}

func (discovery *Discovery) receiver() {
    defer close(discovery.recvChan)

    for {
        var packet DiscoveryPacket

        buf := make([]byte, 1500)

        if n, recvAddr, err := discovery.udpConn.ReadFromUDP(buf); err != nil {
            log.Printf("Discovery.receiver: udpConn.ReadFromUDP: %v\n", err)
            return
        } else {
            packet.addr = recvAddr
            packet.data = buf[:n]
        }

        packet.IP = packet.addr.IP

        discovery.recvChan <- packet
    }
}

func (discovery *Discovery) run(outChan chan DiscoveryPacket) {
    intervalChan := time.Tick(discovery.options.Interval)

    for {
        select {
        case <-intervalChan:
            if err := discovery.send(); err != nil {
                log.Printf("Discovery.Send: %v\n", err)
            } else {
                //log.Printf("Discovery.Send...\n")
            }

        case packet := <-discovery.recvChan:
            //log.Printf("Discovery: recv: %v\n", packet)

            outChan <- packet
        }
    }
}

func (discovery *Discovery) Run() chan DiscoveryPacket {
    outChan := make(chan DiscoveryPacket)

    go discovery.receiver()
    go discovery.run(outChan)

    return outChan
}
