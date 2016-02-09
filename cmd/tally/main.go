package main

import (
    "fmt"
    "github.com/qmsk/e2/client"
    "github.com/qmsk/e2/discovery"
    "github.com/jessevdk/go-flags"
    "log"
)


var options = struct{
    DiscoveryOptions    discovery.Options       `group:"E2 Discovery"`
    ClientOptions       client.Options          `group:"E2 JSON-RPC"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
    if _, err := parser.Parse(); err != nil {
        log.Fatalf("%v\n", err)
    }

    if clientOptions, err := options.ClientOptions.DiscoverClient(options.DiscoveryOptions); err != nil {
        log.Fatalf("Client %#v: Discover %#v: %v\n", options.ClientOptions, options.DiscoveryOptions,err)
    } else if xmlClient, err := clientOptions.XMLClient(); err != nil {
        log.Fatalf("Client %#v: XMLClient: %v", clientOptions, err)
    } else if listenChan, err := xmlClient.Listen(); err != nil {
        log.Fatalf("XMLClient %v: Listen: %v", xmlClient, err)
    } else {
        for system := range listenChan {
            fmt.Printf("\033[H\033[2J")

            for sourceID, source := range system.SrcMgr.SourceCol.Source {
                fmt.Printf("Source %d: %v\n", sourceID, source.Name)

                for screenID, screen := range system.DestMgr.ScreenDestCol.ScreenDest {
                    for _, layer := range screen.LayerCollection.Layer {
                        if layer.LastSrcIdx != sourceID {
                            continue
                        }

                        if layer.PvwMode > 0 {
                            fmt.Printf("\tPreview %d: %v\n", screenID, screen.Name)
                        }
                        if layer.PgmMode > 0 {
                            fmt.Printf("\tProgram %d: %v\n", screenID, screen.Name)
                        }
                    }
                }
            }
        }
    }
}
