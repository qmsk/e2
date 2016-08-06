package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"log"
	"time"
	"regexp"
)

type Options struct {
	clientOptions    client.Options
	discoveryOptions discovery.Options

	IgnoreDest		string	`long:"tally-ignore-dest" value-name:"REGEXP" description:"Ignore matching destinations (case-insensitive regexp)"`
	ignoreDestRegexp	*regexp.Regexp
	ContactName		string	`long:"tally-contact-name" value-name:"NAME" default:"tally" description:"Resolve Input ID from Contact 'tally=\\d' field"`
	contactIDRegexp		*regexp.Regexp
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
	options         Options

	closeChan chan struct{}

	discovery     *discovery.Discovery
	discoveryChan chan discovery.Packet

	sources    sources
	sourceChan chan Source

	state State
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
func (tally *Tally) Register(stateChan chan State) {
	tally.dests[stateChan] = true
}

// mainloop, owns Tally state
func (tally *Tally) Run() error {

	for {
		// update
		tally.apply(tally.state)

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
				log.Printf("Tally: Source %v Error: %v\n", source, err)
			} else {
				log.Printf("Tally: Source %v: Update\n", source)
			}

			source.updated = time.Now()

			if source.closed {
				delete(tally.sources, source.String())
			} else {
				tally.sources[source.String()] = source
			}

			tally.state = tally.update()
		}

		// stopping?
		if tally.closeChan == nil && len(tally.sources) == 0 {
			log.Printf("Tally: stopped")
			return nil
		}
	}
}

func (tally *Tally) getState() State {
	// XXX: unsafe, tally.state access is not atomic
	return tally.state
}

func (tally *Tally) getSources() sources {
	// XXX: unsafe, shared map access
	return tally.sources
}

func (tally *Tally) apply(state State) {
	log.Printf("tally: Update: sources=%d inputs=%d outputs=%d tallys=%d",
		len(tally.sources), len(state.Inputs), len(state.Outputs), len(state.Tally),
	)

	for stateChan, _ := range tally.dests {
		stateChan <- state
	}
}

// Compute new output state from sources
func (tally *Tally) update() State {
	var state = makeState()

	for _, source := range tally.sources {
		if err := source.updateState(&state); err != nil {
			state.setSourceError(source.String(), err)
		}
	}

	state.update()

	return state
}

// Termiante any Run()
func (tally *Tally) Stop() {
	close(tally.closeChan)
}
