package client

import (
    "github.com/qmsk/e2/discovery"
    "fmt"
    "log"
)

// If there is no URL given, use Discovery to find any E2 systems.
// Returns a new Options for the first E2 found.
func (options Options) DiscoverClient(discoveryOptions discovery.Options) (Options, error) {
    if options.Address != "" {
        return options, nil
    } else if discovery, err := discoveryOptions.Discovery(); err != nil {
        return options, err
    } else {
        defer discovery.Stop()

        log.Printf("Discovering systems on %v...\n", discovery)

        for packet := range discovery.Run() {
            options.Address = packet.IP.String()

            log.Printf("Discovered system: %v\n", options.Address)

            return options, nil
        }

        return options, fmt.Errorf("Discovery failed")
    }
}
