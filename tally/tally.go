package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"log"
)

type Options struct {
	clientOptions    client.Options
	discoveryOptions discovery.Options
}

func (options Options) Tally(clientOptions client.Options, discoveryOptions discovery.Options) (*Tally, error) {
	options.clientOptions = clientOptions
	options.discoveryOptions = discoveryOptions

	var tally = Tally{
		options:    options,
		closeChan:  make(chan struct{}),
		sources:    make(map[string]Source),
		sourceChan: make(chan Source),
		dests:      make(map[chan State]bool),
	}

	return &tally, tally.init(options)
}

// Concurrent tally support for multiple sources and destinations
type Tally struct {
	options Options

	closeChan chan struct{}

	discovery     *discovery.Discovery
	discoveryChan chan discovery.Packet

	sources    map[string]Source
	sourceChan chan Source

	dests map[chan State]bool
}

func (tally *Tally) init(options Options) error {

	if discovery, err := options.discoveryOptions.Discovery(); err != nil {
		return fmt.Errorf("discovery:DiscoveryOptions.Discovery: %v", err)
	} else {
		tally.discovery = discovery
		tally.discoveryChan = discovery.Run()
	}

	return nil
}

// Register watcher for state
func (tally *Tally) register(stateChan chan State) {
	tally.dests[stateChan] = true
}

// mainloop, owns Tally state
func (tally *Tally) Run() error {
	for {
		select {
		case <-tally.closeChan:
			log.Printf("Tally: stopping...")

			for _, source := range tally.sources {
				source.close()
			}

			// mark as closed, wait for Sources to finish
			tally.closeChan = nil

		case discoveryPacket, valid := <-tally.discoveryChan:
			if !valid {
				return fmt.Errorf("discovery: %v", tally.discovery.Error())
			} else if clientOptions, err := tally.options.clientOptions.DiscoverOptions(discoveryPacket); err != nil {
				log.Printf("Tally: invalid discovery client options: %v\n", err)
			} else if source, exists := tally.sources[clientOptions.String()]; exists && source.err == nil {
				// already running
			} else if source, err := newSource(tally, clientOptions); err != nil {
				log.Printf("Tally: unable to connect to discovered system: %v\n", err)
			} else {
				log.Printf("Tally: connected to source: %v\n", source)

				tally.sources[clientOptions.String()] = source
			}

		case source := <-tally.sourceChan:
			if err := source.err; err != nil {
				log.Printf("Tally: Source %v Error: %v\n", source, err)
			} else {
				log.Printf("Tally: Source %v: Update\n", source)
			}

			if source.closed {
				delete(tally.sources, source.String())
			} else {
				tally.sources[source.String()] = source
			}

			if err := tally.update(); err != nil {
				return fmt.Errorf("Tally.update: %v\n", err)
			}
		}

		// stopping?
		if tally.closeChan == nil && len(tally.sources) == 0 {
			log.Printf("Tally: stopped")
			return nil
		}
	}
}

// Compute new output state from sources
func (tally *Tally) update() error {
	var state = makeState()

	for _, source := range tally.sources {
		if err := source.updateState(&state); err != nil {
			return err
		}
	}

	if err := state.update(); err != nil {
		return err
	}

	// proagate
	log.Printf("tally: Update: sources=%d inputs=%d outputs=%d tallys=%d",
		len(tally.sources), len(state.Inputs), len(state.Outputs), len(state.Tally),
	)

	for stateChan, _ := range tally.dests {
		stateChan <- state
	}

	return nil
}

// Termiante any Run()
func (tally *Tally) Stop() {
	close(tally.closeChan)
}
