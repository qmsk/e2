package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"log"
	"regexp"
	"sync/atomic"
	"time"
)

type Options struct {
	clientOptions    client.Options
	discoveryOptions discovery.Options

	IgnoreDest       string `long:"tally-ignore-dest" value-name:"REGEXP" description:"Ignore matching destinations (case-insensitive regexp)"`
	ignoreDestRegexp *regexp.Regexp
	ContactName      string `long:"tally-contact-name" value-name:"NAME" default:"tally" description:"Resolve Input ID from Contact 'tally=\\d' field"`
	contactIDRegexp  *regexp.Regexp
}

func (options Options) Tally(clientOptions client.Options, discoveryOptions discovery.Options) (*Tally, error) {
	options.clientOptions = clientOptions
	options.discoveryOptions = discoveryOptions

	if options.IgnoreDest == "" {

		// case-insensitive match
	} else if regexp, err := regexp.Compile("(?i)" + options.IgnoreDest); err != nil {
		return nil, fmt.Errorf("Invalid --tally-ignore-dest=%v: %v", options.IgnoreDest, err)
	} else {
		options.ignoreDestRegexp = regexp
	}

	if regexp, err := regexp.Compile("(?i)" + options.ContactName + "=" + "(\\d+)"); err != nil {
		return nil, fmt.Errorf("Invalid --tally-contact-key=%v: %v", options.ContactName, err)
	} else {
		options.contactIDRegexp = regexp
	}

	var tally = Tally{
		options:    options,
		closeChan:  make(chan struct{}),
		sources:    make(sources),
		sourceChan: make(chan Source),
		dests:      make(map[chan State]bool),
	}

	return &tally, tally.init(options)
}

type sources map[string]Source

// Concurrent tally support for multiple sources and destinations
type Tally struct {
	options Options

	closeChan chan struct{}

	discovery     *discovery.Discovery
	discoveryChan chan discovery.Packet

	sources    sources
	sourceChan chan Source

	state atomic.Value
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
// TODO: optimize to send a pointer to a shared read-only State?
func (tally *Tally) Register(stateChan chan State) {
	tally.dests[stateChan] = true
}

// mainloop, owns Tally state
func (tally *Tally) Run() error {
	// use a nil state to catch anyone trying to change it... :)
	var state *State = &State{}

	for {
		// update
		tally.apply(state)

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
			} else if source, err := newSource(tally, discoveryPacket, clientOptions); err != nil {
				log.Printf("Tally: unable to connect to discovered system: %v\n", err)
			} else {
				log.Printf("Tally: connected to source: %v\n", source)

				tally.sources[clientOptions.String()] = source
			}

		case source := <-tally.sourceChan:
			if err := source.err; err != nil {
				log.Printf("Tally: Source %v Error: %v", source, err)

				tally.sources[source.String()] = source

			} else {
				log.Printf("Tally: Source %v: Update", source)

				source.updated = time.Now()

				tally.sources[source.String()] = source
			}

			state = tally.update()
		}

		// stopping?
		if tally.closeChan == nil {
			var closed = true

			for _, source := range tally.sources {
				if !source.isClosed() {
					closed = false
				}
			}

			if closed {
				log.Printf("Tally: stopped")
				return nil
			}
		}
	}
}

// Return a copy of the current State.
// TODO: optimize to return a pointer to a shared read-only State?
func (tally *Tally) Get() State {
	return *tally.state.Load().(*State)
}

// store and distribute the new State. This is a shared read-only pointer.
func (tally *Tally) apply(state *State) {
	log.Printf("tally: Update: sources=%d inputs=%d outputs=%d tallys=%d",
		len(tally.sources), len(state.Inputs), len(state.Outputs), len(state.Tally),
	)

	// the state
	tally.state.Store(state)

	for stateChan, _ := range tally.dests {
		stateChan <- *state
	}
}

// Compute new output state from sources
func (tally *Tally) update() *State {
	var state = newState()

	for _, source := range tally.sources {
		if err := source.updateState(state); err != nil {
			state.setSourceError(source, err)
		} else {
			state.setSource(source)
		}
	}

	state.update()

	return state
}

// Termiante any Run()
func (tally *Tally) Stop() {
	close(tally.closeChan)
}
