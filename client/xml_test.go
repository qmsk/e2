package client

import (
    "path/filepath"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "testing"
    "time"
)

func testXMLClient(xmlGlob string) *XMLClient {
    streamFiles, err := filepath.Glob(xmlGlob)
    if err != nil {
        panic(fmt.Errorf("filepath.Glob: %v\n", err))
    }

    // setup mock server
    tcpListener, err := net.ListenTCP("tcp", nil)
    if err != nil {
        panic(err)
    }

    go func(tcpListener *net.TCPListener) {
        for {
            tcpConn, err := tcpListener.AcceptTCP()
            if err != nil {
                panic(err)
            }

            // write each .xml file into the testConn end of the pipe, for XMLClient to read
            go func(tcpConn *net.TCPConn) {
                defer tcpConn.CloseWrite()

                for _, xmlPath := range streamFiles {
                    xmlFile, err := os.Open(xmlPath)
                    if err != nil {
                        panic(fmt.Errorf("os.Open: %v\n", err))
                    }
                    defer xmlFile.Close()

                    if write, err := io.Copy(tcpConn, xmlFile); err != nil {
                        panic(fmt.Errorf("io.Copy: %v\n", err))
                    } else {
                        log.Printf("Send %v: %d bytes\n", xmlPath, write)
                    }

                    time.Sleep(1 * time.Second)
                }

                log.Printf("Send: done\n")
            }(tcpConn)

            go func(tcpConn *net.TCPConn) {
                defer tcpConn.Close()

                for {
                    buf := make([]byte, 1500)

                    if read, err := tcpConn.Read(buf); err != nil {
                        log.Printf("Recv error: %v\n", err)
                        break
                    } else {
                        log.Printf("Recv %d bytes: %#v\n", read, string(buf[:read]))
                    }
                }

                log.Printf("Recv: done\n")
            }(tcpConn)
        }
    }(tcpListener)

    // connect client to test server
    listenAddr := tcpListener.Addr().(*net.TCPAddr)

    tcpConn, err := net.DialTCP("tcp", nil, listenAddr)
    if err != nil {
        panic(err)
    }

    xmlClient := XMLClient{
        timeout:    10 * time.Second,
        conn:       tcpConn,
    }

    return &xmlClient
}

func TestXmlRead(t *testing.T) {
    xmlClient := testXMLClient("./test-xml/test1-*.xml")

    listenChan, err := xmlClient.Listen()
    if err != nil {
        t.Fatalf("xmlClient.Listen: %v", err)
    }

    // read System state updates, and updates this to be the final System state after all cumulative updates
    var listenSystem System

    for listenSystem = range listenChan {
        // read it
        _ = listenSystem.String()
    }

    log.Printf("End of Listen\n")

    // check resulting system state
    if source0, exists := listenSystem.SrcMgr.SourceCol[0]; !exists {
        t.Errorf("Source #0 does not exist")
    } else {
        t.Logf("Source #0: %#v\n", source0)

        if source0.Name != "PC 1-1" {
            t.Errorf("Source #0 Name: %v", source0.Name)
        }
    }

    if screen0, exists := listenSystem.DestMgr.ScreenDestCol[0]; !exists {
        t.Errorf("Screen #0 does not exist")
    } else {
        t.Logf("Screen #0: %#v\n", screen0)

        if screen0.Name != "ScreenDest1" {
            t.Errorf("Screen #0 Name: %v", screen0.Name)
        }

        if screen0.IsActive != 1 {
            t.Errorf("Screen #0: IsActive=%v", screen0.IsActive)
        }

        if layer0, exists := screen0.LayerCollection[0]; !exists {
            t.Errorf("Layer #0 does not exist")
        } else {
            t.Logf("Layer #0: %#v\n", layer0)

            if layer0.LastSrcIdx != 0 {
                t.Errorf("Layer #0: LastSrcIdx=%v", layer0.LastSrcIdx)
            }

            if layer0.PgmMode != 1 || layer0.PvwMode != 0 {
                t.Errorf("Layer #0: PgmMode=%v PvwMode=%v", layer0.PgmMode, layer0.PvwMode)
            }
        }

        if layer1, exists := screen0.LayerCollection[1]; !exists {
            t.Errorf("Layer #1 does not exist")
        } else {
            t.Logf("Layer #1: %#v\n", layer1)

            if layer1.LastSrcIdx != 0 {
                t.Errorf("Layer #1: LastSrcIdx=%v", layer1.LastSrcIdx)
            }

            if layer1.PgmMode != 0 || layer1.PvwMode != 1 {
                t.Errorf("Layer #0: PgmMode=%v PvwMode=%v", layer1.PgmMode, layer1.PvwMode)
            }
        }
    }
}
