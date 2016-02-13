package client

import (
    "path/filepath"
    "os"
    "sync"
    "testing"
    "encoding/xml"
)

const testXmlFiles = "./test-xml/test1-*.xml"

func TestXmlRead(t *testing.T) {
    var wg sync.WaitGroup

    listenChan := make(chan System)
    var listenSystem System

    wg.Add(1)
    go func() {
        defer wg.Done()

        for listenSystem = range listenChan {
            // read it
            _ = listenSystem.String()
        }
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(listenChan)

        var system System

        streamFiles, err := filepath.Glob(testXmlFiles)
        if err != nil {
            t.Fatalf("filepath.Glob: %v\n", err)
        }

        for _, xmlPath := range streamFiles {
            xmlFile, err := os.Open(xmlPath)
            if err != nil {
                t.Fatalf("os.Open: %v\n", err)
            }

            packet := xmlPacket{xmlResponse: xmlResponse{System: &system}}

            if err := xml.NewDecoder(xmlFile).Decode(&packet); err != nil {
                t.Fatalf("XML Decode: %v\n", err)
            }

            listenChan <- system
        }
    }()

    wg.Wait()

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
