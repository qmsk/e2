package tally

import (
    "github.com/qmsk/e2/client"
    "github.com/qmsk/e2/discovery"
    "fmt"
    "log"
)

type Options struct {
    clientOptions       client.Options
    discoveryOptions    discovery.Options
}

func (options Options) Tally(clientOptions client.Options, discoveryOptions discovery.Options) (*Tally, error) {
    options.clientOptions = clientOptions
    options.discoveryOptions = discoveryOptions

    var tally Tally

    return &tally, tally.start(options)
}

// Concurrent tally support for multiple sources and destinations
type Tally struct {
    options         Options

    discovery       *discovery.Discovery
    discoveryChan   chan discovery.Packet

    /* run() state */
    // active systems
    sources         map[string]Source

    // updates to sources
    sourceChan      chan Source
}

func (tally *Tally) start(options Options) error {
    tally.options = options
    tally.sources = make(map[string]Source)
    tally.sourceChan = make(chan Source)

    if discovery, err := options.discoveryOptions.Discovery(); err != nil {
        return fmt.Errorf("discovery:DiscoveryOptions.Discovery: %v", err)
    } else {
        tally.discovery = discovery
        tally.discoveryChan = discovery.Run()
    }

    return nil
}

// mainloop, owns Tally state
func (tally *Tally) Run() error {
    for {
        select {
        case discoveryPacket := <-tally.discoveryChan:
            if clientOptions, err := tally.options.clientOptions.DiscoverOptions(discoveryPacket); err != nil {
                log.Printf("Tally: invalid discovery client options: %v\n", err)
            } else if _, exists := tally.sources[clientOptions.String()]; exists {
                // already known
            } else if source, err := newSource(tally, clientOptions); err != nil {
                log.Printf("Tally: unable to connect to discovered system: %v\n", err)
            } else {
                log.Printf("Tally: connected to new source: %v\n", source)

                tally.sources[clientOptions.String()] = source
            }

        case source := <-tally.sourceChan:
            if err := source.err; err != nil {
                log.Printf("Tally: Source %v Error: %v\n", source, err)

                delete(tally.sources, source.String())
            } else {
                log.Printf("Tally: Source %v: Update\n", source)

                tally.sources[source.String()] = source
            }

            if err := tally.update(); err != nil {
                return fmt.Errorf("Tally.update: %v\n", err)
            }
        }
    }
}

// Compute new output state from sources
func (tally *Tally) update() error {
    return nil
}
